# Licensed under the Apache License, Version 2.0 (the “License”);
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#  http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an “AS IS” BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

################################################################################


import json
from argparse import ArgumentParser, Namespace

import pyarrow as pa
import pyarrow.parquet as pq

from transformer import AllowLicenseStatusTransformer, DenyLicenseStatusTransformer

def validate_columns(table: pa.Table, required: list[str]) -> None:
    """
    Check if required columns exist in the table
    :param table: table
    :param required: list of required columns
    :return: None
    """
    columns = table.schema.names
    result = True
    for r in required:
        if r not in columns:
            result = False
            break
    if not result:
        raise Exception(
            f"Not all required columns are present in the table - " f"required {required}, present {columns}"
        )

LICENSE_SELECT_PARAMS = "license_select_params"

shortname = "lc"
CLI_PREFIX = f"{shortname}_"

LICENSE_COLUMN_NAME_KEY = "license_column_name"
LICENSE_COLUMN_NAME_CLI_KEY = f"{CLI_PREFIX}{LICENSE_COLUMN_NAME_KEY}"

DENY_LICENSES_KEY = "deny_licenses"
DENY_LICENSES_CLI_KEY = f"{CLI_PREFIX}{DENY_LICENSES_KEY}"

LICENSES_FILE_KEY = "licenses_file"
LICENSES_FILE_CLI_KEY = f"{CLI_PREFIX}{LICENSES_FILE_KEY}"

ALLOW_NO_LICENSE_KEY = "allow_no_license"
ALLOW_NO_LICENSE_CLI_KEY = f"{CLI_PREFIX}{ALLOW_NO_LICENSE_KEY}"

LICENSE_COLUMN_DEFAULT = "license"
LICENSES_KEY = "licenses"

def _get_supported_licenses(license_file: str, data_access: DataAccess) -> list[str]:
    logger.info(f"Getting supported licenses from file {license_file}")
    licenses_list = None
    try:
        licenses_list_json, _ = data_access.get_file(license_file)
        licenses_list = json.loads(licenses_list_json.decode("utf-8"))
        logger.info(f"Read a list of {len(licenses_list)} licenses.")
    except Exception as e:
        logger.error(f"Failed to read file: {license_file} due to {e}")
    return licenses_list

    """It can be used to select the rows/records of data with licenses
    matching those in the approved/deny list. It adds a new column: `license_status`
    to indicate the selected/denied licenses.

       config: dictionary of configuration data
                license_select_params: A dictionary with the following keys.
                    license_column_name - The name of the column with license, default: 'licence'.
                    allow_no_license - Allows to select rows with no license. default: False
                    licenses - A list of licenses
                    deny_licenses - if selected, the the licenses list is used as deny list, default: False
        Example:
                config = {
                    "license_select_params": {
                        "license_column_name": "license",
                        "allow_no_license": False,
                        "licenses": ["MIT", "Apache 2.0"],
                        "deny_licenses": False
                        }
                }
    """

try:
    license_select = getenv(LICENSE_SELECT_PARAMS, None)
    license_column = license_select.get(LICENSE_COLUMN_NAME_KEY, LICENSE_COLUMN_DEFAULT)
    allow_no_license = license_select.get(ALLOW_NO_LICENSE_KEY, False)
    licenses = license_select.get(LICENSES_KEY, None)
    if not licenses or not isinstance(licenses, list):
        raise ValueError("license list not found.")
    deny = license_select.get(DENY_LICENSES_KEY, False)
    print(f"LICENSE_SELECT_PARAMS: {license_select}")
except Exception as e:
    raise ValueError(f"Invalid Argument: cannot create LicenseSelectTransform object: {e}.")

if not deny:
    transformer = AllowLicenseStatusTransformer(
        license_column=license_column,
        allow_no_license=allow_no_license,
        licenses=licenses,
    )
else:
    transformer = DenyLicenseStatusTransformer(
        license_column=license_column,
        allow_no_license=allow_no_license,
        licenses=licenses,
    )

def transform(table: pa.Table, file_name: str = None) -> tuple[list[pa.Table], dict]:
    """
    Transforms input tables by adding a boolean `license_status` column
    indicating whether the license is approved/denied.
    """
    validate_columns(table=table, required=[license_column])
    new_table = transformer.transform(table)
    return [new_table], {}

try:
    print(f"Reading in parquet file {sys.argv[1]}")
    table = pq.read_table(sys.argv[1])
except Exception as e:
    print(f"Error reading table: {e}", file=sys.stderr)
    exit(1)
print(f"Done Reading in parquet file {sys.argv[1]}")
 
print(f"input table has {table.num_rows} rows")
# Transform the table
table_list, metadata = transform(table, sys.argv[1])
print(f"\noutput table has {table_list[0].num_rows} rows")
print(f"output metadata : {metadata}")
pq.write_table(table_list[0], sys.argv[2])
