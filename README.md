# httpwr

[![Build Status](https://img.shields.io/github/actions/workflow/status/samuelsih/httpwr/go.yaml?branch=master&style=for-the-badge)](https://github.com/samuelsih/httpwr/actions?workflow=build)
![Coverage](https://github.com/samuelsih/httpwr/blob/master/badge.svg)

`httpwr` is an extended and modified version of this [repository](https://github.com/caarlos0/httperr).

The idea is still the same, that is to force the return type to the http handler, so we can minimize the risk of putting return state.

```go
func someHandler(w http.ResponseWriter, r *http.Request) {
	err := doSomethingThatReturnError()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		// forgot to return here
	}

	doJob()
}
```

Basic usages:

1. Return error from handler

   ```go
   func someFunction() error {
    return errors.New("error messages")
   }

   func main() {
   	router := http.NewServeMux()

   	router.Handle("/test", httpwr.F(func(w http.ResponseWriter, r *http.Request) error {
   		err := someFunction()
   		return err
   	}))
   }
   ```

   ```json
   {
     "status": 500,
     "error": "error messages"
   }
   ```

2. Return error with status code

   ```go
   func someFunction() error {
    return errors.New("error messages")
   }

   func main() {
   	router := http.NewServeMux()

   	router.Handle("/test", httpwr.F(func(w http.ResponseWriter, r *http.Request) error {
   		err := someFunction()
   		return httpwr.Wrap(http.StatusBadRequest, err)
   	}))
   }
   ```

   ```json
   {
     "status": 400,
     "error": "error messages"
   }
   ```

3. Return error with `errorf`:

   ```go
   func main() {
   	router := http.NewServeMux()

      router.Handle("/test", httpwr.F(func(w http.ResponseWriter, r *http.Request) error {
		 if somethingWrong {
         return httpwr.Errorf(
            http.StatusBadRequest, 
            "something wrong: %v", somethingWrong,
         )
       }
   	}))
   }
   ```

   ```json
   {
     "status": 400,
     "error": "something wrong: i dont know message"
   }
   ```

4. No Error in handler? You can return `httpwr.OK`

   ```go
   func main() {
   	router := http.NewServeMux()

   	router.Handle("/test", httpwr.F(func(w http.ResponseWriter, r *http.Request) error {
   		return httpwr.OK(http.StatusOK, "all good")
   	}))
   }
   ```

   ```json
   {
     "status": 200,
     "msg": "all good"
   }
   ```

5. Want to return OK with Data? Use `httpwr.OKWithData`

   ```go
   func main() {
   	router := http.NewServeMux()

   	router.Handle("/test", httpwr.F(func(w http.ResponseWriter, r *http.Request) error {
   		data := M{"some": "data"}
         return httpwr.OKWithData(http.StatusOK, "all good", data)
   	}))
   }
   ```

   ```json
   {
     "status": 200,
     "msg": "all good"
     "data": {
        "some": "data",
     }
   }
   ```

   `httpwr.M` is an alias for `map[string]any`
