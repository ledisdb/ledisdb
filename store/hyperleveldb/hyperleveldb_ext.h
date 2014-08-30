// +build hyperleveldb

#ifndef HYPERLEVELDB_EXT_H
#define HYPERLEVELDB_EXT_H

#ifdef __cplusplus
extern "C" {
#endif

#include "hyperleveldb/c.h"


// /* Returns NULL if not found. Otherwise stores the value in **valptr.
//    Stores the length of the value in *vallen. 
//    Returns a context must be later to free*/
// extern void* hyperleveldb_get_ext(
//     leveldb_t* db,
//     const leveldb_readoptions_t* options,
//     const char* key, size_t keylen,
//     char** valptr,
//     size_t* vallen,
//     char** errptr);

// // Free context returns by hyperleveldb_get_ext
// extern void hyperleveldb_get_free_ext(void* context);


// Below iterator functions like leveldb iterator but returns valid status for iterator
extern unsigned char hyperleveldb_iter_seek_to_first_ext(leveldb_iterator_t*);
extern unsigned char hyperleveldb_iter_seek_to_last_ext(leveldb_iterator_t*);
extern unsigned char hyperleveldb_iter_seek_ext(leveldb_iterator_t*, const char* k, size_t klen);
extern unsigned char hyperleveldb_iter_next_ext(leveldb_iterator_t*);
extern unsigned char hyperleveldb_iter_prev_ext(leveldb_iterator_t*);


#ifdef __cplusplus
}
#endif

#endif