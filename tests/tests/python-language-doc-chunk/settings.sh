api=workqueue

expected=("Transforming one table with 1 rows" "Done with nfiles=1 nrows=88")
NUM_DESIRED_OUTPUTS=0

up_args='<(gunzip -c "$TEST_PATH"/pail/test-data/input/test1.parquet.gz)'
