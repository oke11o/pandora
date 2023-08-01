package scenario

const exampleAmmoFile = `
variables:
  hostname: localhost

variablesources:
  - type: "file/csv"
    name: "users_src"
    file: "_files/users.csv"
    mapping:
      users: [ "user_id", "name", "pass", "created_at" ]
  - type: "file/json"
    name: "filter_src"
    file: "_files/filter.json"

requests:
  - name: "auth_req"
    uri: '/auth'
    method: POST
    headers:
      Useragent: Tank
      ContentType: "application/json"
      Hostname: "{{hostname}}"
    tag: auth
    body: '{"user_name": {{source.users_src.users[next].name}}, "user_pass": {{source.users_src.users[next].pass}} }'
    templater: text
    postprocessors:
      - type: var/header
        mapping:
          httpAuthorization: "Http-Authorization"
          contentType: "Content-Type|lower"
      - type: 'var/jsonpath'
        mapping:
          token: "$.data.authToken"

  - name: list_req
    preprocessors:
      - type: prepare
        mapping:
          filter: source.filter_src.list[rand]
    uri: '/list/?{{filter|query}}'
    method: GET
    headers:
      Useragent: "Tank"
      ContentType: "application/json"
      Hostname: "{{hostname}}"
      Authorization: "Bearer {{request.auth_req.token}}"
    tag: list
    postprocessors:
      - type: var/jsonpath
        mapping:
          items: $.data.items

  - name: order_req
    preprocessors:
      - type: prepare
        mapping:
          item: list_req.items.items[rand]
    uri: '/order'
    tag: order	
    method: POST
    headers:
      Useragent: "Tank"
      ContentType: "application/json"
      Hostname: "{{hostname}}"
      Authorization: "Bearer {{request.auth_req.token}}"
    body: "{}"
    postprocessors:
      - type: var/jsonpath
        mapping:
          delivery_id: $.data.delivery_id

scenarios:
  - name: scenario1
    weight: 50
    minwaitingtime: 1000
    shoot: [
      auth(1),
      sleep(100),
      list(1),
      sleep(100),
      order(3)
    ]
  - name: scenario2
    weight: 10
    minwaitingtime: 1000
    shoot: [
      auth(1),
      sleep(100),
      list(1),
      sleep(100),
      order(3)
    ]
`
