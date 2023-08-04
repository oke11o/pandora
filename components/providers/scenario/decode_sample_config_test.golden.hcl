variables = {
  hostname = "localhost"
}

variablesource "users_src" "file/csv" {
  file             = "_files/users.csv"
  fields           = ["user_id", "name", "pass", "created_at"]
  skip_header      = false
  header_as_fields = false
}
variablesource "filter_src" "file/json" {
  file             = "_files/filter.json"
  fields           = null
  skip_header      = false
  header_as_fields = false
}

request "auth_req" {
  method = "POST"
  headers = {
    ContentType = "application/json"
    Hostname    = "{{hostname}}"
    Useragent   = "Tank"
  }
  tag  = "auth"
  body = "{\"user_name\": {{source.users_src.users[next].name}}, \"user_pass\": {{source.users_src.users[next].pass}} }"
  uri  = "/auth"

  preprocessor "" {
    variables = null
  }

  postprocessor "var/header" {
    mapping = {
      contentType       = "Content-Type|lower"
      httpAuthorization = "Http-Authorization"
    }
  }
  postprocessor "var/jsonpath" {
    mapping = {
      token = "$.data.authToken"
    }
  }

  templater = "text"
}
request "list_req" {
  method = "GET"
  headers = {
    Authorization = "Bearer {{request.auth_req.token}}"
    ContentType   = "application/json"
    Hostname      = "{{hostname}}"
    Useragent     = "Tank"
  }
  tag = "list"
  uri = "/list/?{{filter|query}}"

  preprocessor "prepare" {
    variables = {
      filter = "source.filter_src.list[rand]"
    }
  }

  postprocessor "var/jsonpath" {
    mapping = {
      items = "$.data.items"
    }
  }

  templater = ""
}
request "order_req" {
  method = "POST"
  headers = {
    Authorization = "Bearer {{request.auth_req.token}}"
    ContentType   = "application/json"
    Hostname      = "{{hostname}}"
    Useragent     = "Tank"
  }
  tag  = "order"
  body = "{}"
  uri  = "/order"

  preprocessor "prepare" {
    variables = {
      item = "list_req.items.items[rand]"
    }
  }

  postprocessor "var/jsonpath" {
    mapping = {
      delivery_id = "$.data.delivery_id"
    }
  }

  templater = ""
}

scenario "scenario1" {
  weight           = 50
  min_waiting_time = 1000
  shoot            = ["auth(1)", "sleep(100)", "list(1)", "sleep(100)", "order(3)"]
}
scenario "scenario2" {
  weight           = 10
  min_waiting_time = 1000
  shoot            = ["auth(1)", "sleep(100)", "list(1)", "sleep(100)", "order(3)"]
}