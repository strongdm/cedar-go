use std::env;
use std::fs;
use std::path::Path;

use cedar_policy::{
    Context, Entities, EntityId, EntityTypeName, EntityUid, PolicySet, Request, Schema,
    ValidationMode, Validator,
};
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Deserialize)]
struct TestFile {
    schema: String,
    policies: String,
    entities: String,
    requests: Vec<TestRequest>,
}

#[derive(Deserialize)]
struct TestRequest {
    description: String,
    principal: EntityRef,
    action: EntityRef,
    resource: EntityRef,
    context: Value,
    #[serde(rename = "validateRequest")]
    validate_request: bool,
}

#[derive(Deserialize)]
struct EntityRef {
    #[serde(rename = "type")]
    entity_type: String,
    id: String,
}

#[derive(Serialize)]
struct ValidationResult {
    #[serde(rename = "policyValidation")]
    policy_validation: ModeResult,
    #[serde(rename = "entityValidation")]
    entity_validation: ModeResult,
    #[serde(rename = "requestValidation")]
    request_validation: Vec<RequestResult>,
}

#[derive(Serialize)]
struct ModeResult {
    strict: bool,
    permissive: bool,
    #[serde(rename = "strictErrors", skip_serializing_if = "Vec::is_empty")]
    strict_errors: Vec<String>,
    #[serde(rename = "permissiveErrors", skip_serializing_if = "Vec::is_empty")]
    permissive_errors: Vec<String>,
}

#[derive(Serialize)]
struct RequestResult {
    description: String,
    strict: Option<bool>,
    permissive: Option<bool>,
    #[serde(rename = "errors", skip_serializing_if = "Vec::is_empty")]
    errors: Vec<String>,
}

fn make_entity_uid(r: &EntityRef) -> EntityUid {
    EntityUid::from_type_name_and_id(
        r.entity_type.parse::<EntityTypeName>().unwrap(),
        EntityId::new(&r.id),
    )
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!("Usage: {} <test-json-path> <output-json-path>", args[0]);
        std::process::exit(1);
    }

    let test_json_path = &args[1];
    let output_json_path = &args[2];

    let test_json = fs::read_to_string(test_json_path).expect("Failed to read test JSON");
    let test: TestFile = serde_json::from_str(&test_json).expect("Failed to parse test JSON");

    // Resolve file paths relative to the parent of the corpus-tests directory
    let base_dir = Path::new(test_json_path)
        .parent()
        .unwrap()
        .parent()
        .unwrap();

    // Read schema
    let schema_str = fs::read_to_string(base_dir.join(&test.schema)).expect("Failed to read schema");
    let (schema, _warnings) =
        Schema::from_cedarschema_str(&schema_str).expect("Failed to parse schema");

    // Read policies
    let policies_str =
        fs::read_to_string(base_dir.join(&test.policies)).expect("Failed to read policies");
    let policy_set: PolicySet = policies_str.parse().expect("Failed to parse policies");

    // Read entities
    let entities_str =
        fs::read_to_string(base_dir.join(&test.entities)).expect("Failed to read entities");

    // 1. Policy validation (strict and permissive)
    let validator = Validator::new(schema.clone());
    let strict_policy = validator.validate(&policy_set, ValidationMode::Strict);
    let permissive_policy = validator.validate(&policy_set, ValidationMode::Permissive);

    let policy_validation = ModeResult {
        strict: strict_policy.validation_passed(),
        permissive: permissive_policy.validation_passed(),
        strict_errors: strict_policy
            .validation_errors()
            .map(|e| format!("{e}"))
            .collect(),
        permissive_errors: permissive_policy
            .validation_errors()
            .map(|e| format!("{e}"))
            .collect(),
    };

    // 2. Entity validation
    // The Rust Entities API doesn't take a validation mode parameter,
    // so strict and permissive produce the same result.
    let entity_result = Entities::from_json_str(&entities_str, Some(&schema));
    let entity_valid = entity_result.is_ok();
    let entity_errors: Vec<String> = match &entity_result {
        Err(e) => vec![format!("{e}")],
        Ok(_) => vec![],
    };
    let entity_validation = ModeResult {
        strict: entity_valid,
        permissive: entity_valid,
        strict_errors: entity_errors.clone(),
        permissive_errors: entity_errors,
    };

    // 3. Per-request validation
    let mut request_validation = Vec::new();
    for req in &test.requests {
        if !req.validate_request {
            request_validation.push(RequestResult {
                description: req.description.clone(),
                strict: None,
                permissive: None,
                errors: vec![],
            });
            continue;
        }

        let principal = make_entity_uid(&req.principal);
        let action = make_entity_uid(&req.action);
        let resource = make_entity_uid(&req.resource);

        // Validate request: build context with schema, then build request with schema.
        // The Rust Request API doesn't take a validation mode parameter,
        // so strict and permissive produce the same result.
        let result = (|| -> Result<(), Box<dyn std::error::Error>> {
            let context =
                Context::from_json_value(req.context.clone(), Some((&schema, &action)))?;
            Request::new(principal, action, resource, context, Some(&schema))?;
            Ok(())
        })();
        let valid = result.is_ok();
        let errors: Vec<String> = match &result {
            Err(e) => vec![format!("{e}")],
            Ok(_) => vec![],
        };

        request_validation.push(RequestResult {
            description: req.description.clone(),
            strict: Some(valid),
            permissive: Some(valid),
            errors,
        });
    }

    let result = ValidationResult {
        policy_validation,
        entity_validation,
        request_validation,
    };

    let output = serde_json::to_string_pretty(&result).expect("Failed to serialize result");
    fs::write(output_json_path, output).expect("Failed to write output");
}
