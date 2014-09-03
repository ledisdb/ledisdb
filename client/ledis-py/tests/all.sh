dbs=(leveldb rocksdb hyperleveldb goleveldb boltdb lmdb)
for db in "${dbs[@]}"
do 
    ledis-server -db_name=$db &
    py.test
    killall ledis-server
done
