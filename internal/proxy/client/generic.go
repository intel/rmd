package proxyclient

// Enforce simplifies enforcing data in a more generic way
func Enforce(module string, params map[string]interface{}) (string, error) {

	params["RMDMODULE"] = module
	var result string
	err := Client.Call("Proxy.Enforce", params, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

// Release simplifies releasing data in a more generic way
func Release(module string, params map[string]interface{}) error {

	params["RMDMODULE"] = module
	err := Client.Call("Proxy.Release", params, nil)
	if err != nil {
		return err
	}
	return nil
}
