variables = {
  hostname = "localhost"
}

variable_source "users" "file/csv" {
  file             = "files/users.csv"
  fields           = ["user_id", "name", "pass"]
  skip_header      = true
  header_as_fields = false
}
variable_source "filter_src" "file/json" {
  file             = "files/filter.json"
  fields           = null
  skip_header      = false
  header_as_fields = false
}

request "auth_req" {
  method = "POST"
  headers = {
    Content-Type = "application/json"
    Useragent    = "Tank"
  }
  tag  = "auth"
  body = "{\"user_id\":  {{.preprocessor.user_id}}}"
  uri  = "/auth"

  preprocessor "" {
    variables = {
      user_id = "source.users[0].user_id"
    }
  }

  postprocessor "var/header" {
    mapping = {
      Content-Type      = "Content-Type|upper"
      httpAuthorization = "Http-Authorization"
    }
  }
  postprocessor "var/jsonpath" {
    mapping = {
      token = "$.auth_key"
    }
  }

  templater = "text"
}
request "list_req" {
  method = "GET"
  headers = {
    Authorization = "Bearer {{.request.auth_req.token}}"
    Content-Type  = "application/json"
    Useragent     = "Tank"
  }
  tag = "list"
  uri = "/list"

  preprocessor "" {
    variables = null
  }

  postprocessor "var/jsonpath" {
    mapping = {
      item_id = "$.items[0]"
      items   = "$.items"
    }
  }

  templater = ""
}
request "item_req" {
  method = "POST"
  headers = {
    Authorization = "Bearer {{.request.auth_req.token}}"
    Content-Type  = "application/json"
    Useragent     = "Tank"
  }
  tag  = "item_req"
  body = "{\"item_id\": {{.preprocessor.item}}}"
  uri  = "/item"

  preprocessor "" {
    variables = {
      item = "request.list_req.items[3]"
    }
  }
  templater = ""
}

scenario "scenario1" {
  weight           = 50
  min_waiting_time = 500
  shoot            = ["auth_req(1)", "sleep(100)", "list_req(1)", "sleep(100)", "item_req(3)"]
}
scenario "scenario2" {
  weight           = 50
  min_waiting_time = 500
  shoot            = ["auth_req(1)", "sleep(100)", "list_req(1)", "sleep(100)", "item_req(2)"]
}
