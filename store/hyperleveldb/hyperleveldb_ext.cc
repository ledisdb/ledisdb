// +build hyperleveldb

#include "hyperleveldb_ext.h"

#include <stdlib.h>
#include <string>

#include "hyperleveldb/db.h"

using namespace leveldb;

extern "C" {

static bool SaveError(char** errptr, const Status& s) {
  assert(errptr != NULL);
  if (s.ok()) {
    return false;
  } else if (*errptr == NULL) {
    *errptr = strdup(s.ToString().c_str());
  } else {
    free(*errptr);
    *errptr = strdup(s.ToString().c_str());
  }
  return true;
}

void* hyperleveldb_get_ext(
    leveldb_t* db,
    const leveldb_readoptions_t* options,
    const char* key, size_t keylen,
    char** valptr,
    size_t* vallen,
    char** errptr) {

    std::string *tmp = new(std::string);

    //very tricky, maybe changed with c++ leveldb upgrade
    Status s = (*(DB**)db)->Get(*(ReadOptions*)options, Slice(key, keylen), tmp);

    if (s.ok()) {
        *valptr = (char*)tmp->data();
        *vallen = tmp->size();
    } else {
        delete(tmp);
        tmp = NULL;
        *valptr = NULL;
        *vallen = 0;
        if (!s.IsNotFound()) {
            SaveError(errptr, s);
        }
    }
    return tmp;
}

void hyperleveldb_get_free_ext(void* context) {
    std::string* s = (std::string*)context;

    delete(s);
}


unsigned char hyperleveldb_iter_seek_to_first_ext(leveldb_iterator_t* iter) {
    leveldb_iter_seek_to_first(iter);
    return leveldb_iter_valid(iter);
}

unsigned char hyperleveldb_iter_seek_to_last_ext(leveldb_iterator_t* iter) {
    leveldb_iter_seek_to_last(iter);
    return leveldb_iter_valid(iter);
}

unsigned char hyperleveldb_iter_seek_ext(leveldb_iterator_t* iter, const char* k, size_t klen) {
    leveldb_iter_seek(iter, k, klen);
    return leveldb_iter_valid(iter);
}

unsigned char hyperleveldb_iter_next_ext(leveldb_iterator_t* iter) {
    leveldb_iter_next(iter);
    return leveldb_iter_valid(iter);
}

unsigned char hyperleveldb_iter_prev_ext(leveldb_iterator_t* iter) {
    leveldb_iter_prev(iter);
    return leveldb_iter_valid(iter);
}


}