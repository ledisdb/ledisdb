dbs=(leveldb rocksdb goleveldb boltdb lmdb)
for db in "${dbs[@]}"
do 
    killall ledis-server
    ledis-server -db_name=$db &
    py.test
done
