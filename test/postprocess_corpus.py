#!/usr/bin/env python3
"""Filter corpus tests that use unsupported extensions (e.g. offset).

Based on upstream: https://github.com/cedar-policy/cedar-integration-tests/blob/main/scripts/postprocess_corpus.py
Extended to also check the test JSON (request contexts), not just entities files.
"""
import tarfile
import json
import os
import tempfile


def has_offset_extension(content):
    return '"fn": "offset"' in content


def process_corpus_tests(input_tar, output_tar):
    with tempfile.TemporaryDirectory() as temp_dir:
        with tarfile.open(input_tar, "r:gz") as tar:
            tar.extractall(temp_dir, filter="data")

        corpus_dir = os.path.join(temp_dir, "corpus-tests")
        files_to_remove = set()

        for filename in sorted(os.listdir(corpus_dir)):
            if filename.endswith(".json") and not filename.endswith(".entities.json"):
                json_path = os.path.join(corpus_dir, filename)
                should_remove = False

                try:
                    with open(json_path, "r") as f:
                        content = f.read()

                    # Check the test JSON itself (request contexts)
                    if has_offset_extension(content):
                        should_remove = True

                    # Check the entities file
                    if not should_remove:
                        test_data = json.loads(content)
                        if isinstance(test_data, dict):
                            entities_file = test_data.get("entities", "").replace(
                                "corpus-tests/", ""
                            )
                            if entities_file:
                                entities_path = os.path.join(corpus_dir, entities_file)
                                if os.path.exists(entities_path):
                                    with open(entities_path, "r") as f:
                                        if has_offset_extension(f.read()):
                                            should_remove = True

                except json.JSONDecodeError:
                    pass

                if should_remove:
                    base_name = filename.replace(".json", "")
                    files_to_remove.add(filename)
                    files_to_remove.add(f"{base_name}.cedar")
                    files_to_remove.add(f"{base_name}.entities.json")
                    files_to_remove.add(f"{base_name}.cedarschema")

        removed_count = 0
        for filename in sorted(files_to_remove):
            file_path = os.path.join(corpus_dir, filename)
            if os.path.exists(file_path):
                os.remove(file_path)
                removed_count += 1

        print(f"Removed {removed_count} files ({len(files_to_remove) // 4} tests using offset extension)")

        with tarfile.open(output_tar, "w:gz") as tar:
            tar.add(corpus_dir, arcname="corpus-tests")


if __name__ == "__main__":
    process_corpus_tests("corpus-tests.tar.gz", "corpus-tests.tar.gz")
