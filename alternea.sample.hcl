log_level = "info"

server "main" {

  listen = ":8080"

  // read_timeout is the maximum duration for reading the entire
  // request, including the body.
  read_timeout = duration("1m") // optional, default is infinite

  // is the maximum duration before timing out
  // It is reset whenever a new request's header is read.
  write_timeout = duration("30s") // optional, default is infinite

  // idle_timeout is the maximum amount of time to wait for the
  // next request when keep-alives are enabled. If idle_timeout
  // is not set, the value of read_timeout is used. If both are
  // not set, there is no timeout.
  idle_timeout = duration("30s") // optional

  proxy_service "/csv/posts" {
    // defines endpoint settings (required)
    backend {
      // specifies endpoint url
      target_url = "https://jsonplaceholder.typicode.com/posts"

      // determines which codes to consider successful, if the end server responds with one of the listed codes,
      // then alternea continues processing the response data, otherwise the original response is
      // transmitted to the client
      success_http_status_codes = [200] // optional, default [200]
    }

    // this transformer expects data in the form of a JSON object
    // in which property names are interpreted as column names
    // the object property values must be an array in which each element is interpreted as a cell value
    transformer "csv" {
      // use_header set to true to use header as a first line (column names)
      use_header = true // optional, default false

      // delimiter specifies delimiter character
      delimiter = ","  // optional, default ","

      // use_crlf set to true to use \r\n as the line terminator
      use_crlf = false // optional, default false

      // tablifier transforms incoming data into a tabular form ([][]string)
      // required by the "csv" transformer
      tablifier = {
        name = "json" // the only available tablifier for now is json

        // before passing the data to the tablifier, they can be preprocessed
        // by optionally defining "remapper"
        remapper = {
          name = "kazaam"
          spec = fromFile("shift.sample.spec.json")
        }
      }
    }
  }


  static_service "/health-check" {
    // defines the HTTP method by which the client should request this service (involved in route matching)
    method = "GET" // optional, default "GET", available methods are: GET, POST, PUT, PATCH, DELETE

    // response_code sets HTTP response status code
    response_code = 200 // required

    // set_header allows to set HTTP headers, optional
    set_header    = {
      X-My-Custom-Header = "custom header"
    }

    // content specifies response content
    content = "Hello!" // optional
  }

}
