package proxyclient

// Code below is a preparation for future more generic proxy implementation

// GenericCaller is a generic method for calling methods on server side
//
// Pointer to this function is passed as one of configuration parameters to
// ModuleInterface.Initialize()
var GenericCaller = func(name string, params map[string]interface{}) error {
	// TODO Consider filling all params by caller (plugin)
	//      and removal of 'name' argument from function signature
	params["PROXYMODULE"] = "pstate"
	params["PROXYFUNCTION"] = name
	return Client.Call("Proxy.GenericCall", params, nil)
}
