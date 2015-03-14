// +build lua

#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include <stdint.h>
#include <stdio.h>
#include "_cgo_export.h"

#define MT_GOFUNCTION "GoLua.GoFunction"
#define MT_GOINTERFACE "GoLua.GoInterface"

#define GOLUA_DEFAULT_MSGHANDLER "golua_default_msghandler"

static const char GoStateRegistryKey = 'k'; //golua registry key
static const char PanicFIDRegistryKey = 'k';

/* taken from lua5.2 source */
void *testudata(lua_State *L, int ud, const char *tname)
{
	void *p = lua_touserdata(L, ud);
	if (p != NULL)
	{  /* value is a userdata? */
		if (lua_getmetatable(L, ud))
		{  /* does it have a metatable? */
			luaL_getmetatable(L, tname);  /* get correct metatable */
			if (!lua_rawequal(L, -1, -2))  /* not the same? */
				p = NULL;  /* value is a userdata with wrong metatable */
			lua_pop(L, 2);  /* remove both metatables */
			return p;
		}
	}
	return NULL;  /* value is not a userdata with a metatable */
}

int clua_isgofunction(lua_State *L, int n)
{
	return testudata(L, n, MT_GOFUNCTION) != NULL;
}

int clua_isgostruct(lua_State *L, int n)
{
	return testudata(L, n, MT_GOINTERFACE) != NULL;
}

unsigned int* clua_checkgosomething(lua_State* L, int index, const char *desired_metatable)
{
	if (desired_metatable != NULL)
	{
		return testudata(L, index, desired_metatable);
	}
	else
	{
		unsigned int *sid = testudata(L, index, MT_GOFUNCTION);
		if (sid != NULL) return sid;
		return testudata(L, index, MT_GOINTERFACE);
	}
}

GoInterface* clua_getgostate(lua_State* L)
{
	//get gostate from registry entry
	lua_pushlightuserdata(L,(void*)&GoStateRegistryKey);
	lua_gettable(L, LUA_REGISTRYINDEX);
	GoInterface* gip = lua_touserdata(L,-1);
	lua_pop(L,1);
	return gip;
}


//wrapper for callgofunction
int callback_function(lua_State* L)
{
	int r;
	unsigned int *fid = clua_checkgosomething(L, 1, MT_GOFUNCTION);
	GoInterface* gi = clua_getgostate(L);
	//remove the go function from the stack (to present same behavior as lua_CFunctions)
	lua_remove(L,1);
	return golua_callgofunction(*gi, fid!=NULL ? *fid : -1);
}

//wrapper for gchook
int gchook_wrapper(lua_State* L)
{
	//printf("Garbage collection wrapper\n");
	unsigned int* fid = clua_checkgosomething(L, -1, NULL);
	GoInterface* gi = clua_getgostate(L);
	if (fid != NULL)
		return golua_gchook(*gi,*fid);
	return 0;
}

unsigned int clua_togofunction(lua_State* L, int index)
{
	unsigned int *r = clua_checkgosomething(L, index, MT_GOFUNCTION);
	return (r != NULL) ? *r : -1;
}

unsigned int clua_togostruct(lua_State *L, int index)
{
	unsigned int *r = clua_checkgosomething(L, index, MT_GOINTERFACE);
	return (r != NULL) ? *r : -1;
}

void clua_pushgofunction(lua_State* L, unsigned int fid)
{
	unsigned int* fidptr = (unsigned int *)lua_newuserdata(L, sizeof(unsigned int));
	*fidptr = fid;
	luaL_getmetatable(L, MT_GOFUNCTION);
	lua_setmetatable(L, -2);
}

static int callback_c (lua_State* L)
{
	int fid = clua_togofunction(L,lua_upvalueindex(1));
	GoInterface *gi = clua_getgostate(L);
	return golua_callgofunction(*gi,fid);
}

void clua_pushcallback(lua_State* L)
{
	lua_pushcclosure(L,callback_c,1);
}

void clua_pushgostruct(lua_State* L, unsigned int iid)
{
	unsigned int* iidptr = (unsigned int *)lua_newuserdata(L, sizeof(unsigned int));
	*iidptr = iid;
	luaL_getmetatable(L, MT_GOINTERFACE);
	lua_setmetatable(L,-2);
}

int default_panicf(lua_State *L)
{
	const char *s = lua_tostring(L, -1);
	printf("Lua unprotected panic: %s\n", s);
	abort();
}

void clua_setgostate(lua_State* L, GoInterface gi)
{
	lua_atpanic(L, default_panicf);
	lua_pushlightuserdata(L,(void*)&GoStateRegistryKey);
	GoInterface* gip = (GoInterface*)lua_newuserdata(L,sizeof(GoInterface));
	//copy interface value to userdata
	gip->v = gi.v;
	gip->t = gi.t;
	//set into registry table
	lua_settable(L, LUA_REGISTRYINDEX);
}

/* called when lua code attempts to access a field of a published go object */
int interface_index_callback(lua_State *L)
{
	unsigned int *iid = clua_checkgosomething(L, 1, MT_GOINTERFACE);
	if (iid == NULL)
	{
		lua_pushnil(L);
		return 1;
	}

	char *field_name = (char *)lua_tostring(L, 2);
	if (field_name == NULL)
	{
		lua_pushnil(L);
		return 1;
	}

	GoInterface* gi = clua_getgostate(L);

	int r = golua_interface_index_callback(*gi, *iid, field_name);

	if (r < 0)
	{
		lua_error(L);
		return 0;
	}
	else
	{
		return r;
	}
}

/* called when lua code attempts to set a field of a published go object */
int interface_newindex_callback(lua_State *L)
{
	unsigned int *iid = clua_checkgosomething(L, 1, MT_GOINTERFACE);
	if (iid == NULL)
	{
		lua_pushnil(L);
		return 1;
	}

	char *field_name = (char *)lua_tostring(L, 2);
	if (field_name == NULL)
	{
		lua_pushnil(L);
		return 1;
	}

	GoInterface* gi = clua_getgostate(L);

	int r = golua_interface_newindex_callback(*gi, *iid, field_name);

	if (r < 0)
	{
		lua_error(L);
		return 0;
	}
	else
	{
		return r;
	}
}

int panic_msghandler(lua_State *L)
{
	GoInterface* gi = clua_getgostate(L);
	go_panic_msghandler(*gi, (char *)lua_tolstring(L, -1, NULL));
	return 0;
}

void clua_hide_pcall(lua_State *L)
{
	lua_getglobal(L, "pcall");
	lua_setglobal(L, "unsafe_pcall");
	lua_pushnil(L);
	lua_setglobal(L, "pcall");

	lua_getglobal(L, "xpcall");
	lua_setglobal(L, "unsafe_xpcall");
	lua_pushnil(L);
	lua_setglobal(L, "xpcall");
}

void clua_initstate(lua_State* L)
{
	/* create the GoLua.GoFunction metatable */
	luaL_newmetatable(L, MT_GOFUNCTION);

	// gofunction_metatable[__call] = &callback_function
	lua_pushliteral(L,"__call");
	lua_pushcfunction(L,&callback_function);
	lua_settable(L,-3);

	// gofunction_metatable[__gc] = &gchook_wrapper
	lua_pushliteral(L,"__gc");
	lua_pushcfunction(L,&gchook_wrapper);
	lua_settable(L,-3);
	lua_pop(L,1);

	luaL_newmetatable(L, MT_GOINTERFACE);

	// gointerface_metatable[__gc] = &gchook_wrapper
	lua_pushliteral(L, "__gc");
	lua_pushcfunction(L, &gchook_wrapper);
	lua_settable(L, -3);

	// gointerface_metatable[__index] = &interface_index_callback
	lua_pushliteral(L, "__index");
	lua_pushcfunction(L, &interface_index_callback);
	lua_settable(L, -3);

	// gointerface_metatable[__newindex] = &interface_newindex_callback
	lua_pushliteral(L, "__newindex");
	lua_pushcfunction(L, &interface_newindex_callback);
	lua_settable(L, -3);

	lua_register(L, GOLUA_DEFAULT_MSGHANDLER, &panic_msghandler);
	lua_pop(L, 1);
}


int callback_panicf(lua_State* L)
{
	lua_pushlightuserdata(L,(void*)&PanicFIDRegistryKey);
	lua_gettable(L,LUA_REGISTRYINDEX);
	unsigned int fid = lua_tointeger(L,-1);
	lua_pop(L,1);
	GoInterface* gi = clua_getgostate(L);
	return golua_callpanicfunction(*gi,fid);

}

//TODO: currently setting garbage when panicf set to null
GoInterface clua_atpanic(lua_State* L, unsigned int panicf_id)
{
	//get old panicfid
	unsigned int old_id;
	lua_pushlightuserdata(L, (void*)&PanicFIDRegistryKey);
	lua_gettable(L,LUA_REGISTRYINDEX);
	if(lua_isnil(L, -1) == 0)
		old_id = lua_tointeger(L,-1);
	lua_pop(L, 1);

	//set registry key for function id of go panic function
	lua_pushlightuserdata(L, (void*)&PanicFIDRegistryKey);
	//push id value
	lua_pushinteger(L, panicf_id);
	//set into registry table
	lua_settable(L, LUA_REGISTRYINDEX);

	//now set the panic function
	lua_CFunction pf = lua_atpanic(L,&callback_panicf);
	//make a GoInterface with a wrapped C panicf or the original go panicf
	if(pf == &callback_panicf)
	{
		return golua_idtointerface(old_id);
	}
	else
	{
		//TODO: technically UB, function ptr -> non function ptr
		return golua_cfunctiontointerface((GoUintptr *)pf);
	}
}

int clua_callluacfunc(lua_State* L, lua_CFunction f)
{
	return f(L);
}

void* allocwrapper(void* ud, void *ptr, size_t osize, size_t nsize)
{
	return (void*)golua_callallocf((GoUintptr)ud,(GoUintptr)ptr,osize,nsize);
}

lua_State* clua_newstate(void* goallocf)
{
	return lua_newstate(&allocwrapper,goallocf);
}

void clua_setallocf(lua_State* L, void* goallocf)
{
	lua_setallocf(L,&allocwrapper,goallocf);
}

void clua_openbase(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_base);
	lua_pushstring(L,"");
	lua_call(L, 1, 0);
	clua_hide_pcall(L);
}

void clua_openio(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_io);
	lua_pushstring(L,"io");
	lua_call(L, 1, 0);
}

void clua_openmath(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_math);
	lua_pushstring(L,"math");
	lua_call(L, 1, 0);
}

void clua_openpackage(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_package);
	lua_pushstring(L,"package");
	lua_call(L, 1, 0);
}

void clua_openstring(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_string);
	lua_pushstring(L,"string");
	lua_call(L, 1, 0);
}

void clua_opentable(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_table);
	lua_pushstring(L,"table");
	lua_call(L, 1, 0);
}

void clua_openos(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_os);
	lua_pushstring(L,"os");
	lua_call(L, 1, 0);
}

void clua_hook_function(lua_State *L, lua_Debug *ar)
{
	lua_checkstack(L, 2);
	lua_pushstring(L, "Lua execution quantum exceeded");
	lua_error(L);
}

void clua_setexecutionlimit(lua_State* L, int n)
{
	lua_sethook(L, &clua_hook_function, LUA_MASKCOUNT, n);
}

LUALIB_API int (luaopen_cjson) (lua_State *L);
LUALIB_API int (luaopen_struct) (lua_State *L);
LUALIB_API int (luaopen_cmsgpack) (lua_State *L);

void clua_opencjson(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_cjson);
	lua_pushstring(L,"cjson");
	lua_call(L, 1, 0);
}

void clua_openstruct(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_struct);
	lua_pushstring(L,"struct");
	lua_call(L, 1, 0);
}

void clua_opencmsgpack(lua_State* L)
{
	lua_pushcfunction(L,&luaopen_cmsgpack);
	lua_pushstring(L,"cmsgpack");
	lua_call(L, 1, 0);
}
