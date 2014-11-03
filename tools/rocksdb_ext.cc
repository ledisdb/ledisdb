#include <rocksdb/c.h>
#include <rocksdb/db.h>

#include <string>

using namespace rocksdb;

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

void* rocksdb_get_ext(
    rocksdb_t* db,
    const rocksdb_readoptions_t* options,
    const char* key, size_t keylen,
    char** valptr,
    size_t* vallen,
    char** errptr) {

    std::string *tmp = new(std::string);

    //very tricky, maybe changed with c++ rocksdb upgrade
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

void rocksdb_get_free_ext(void* context) {
    std::string* s = (std::string*)context;

    delete(s);
}

}