---
title: RPS на инстанс
description: Настройка позволяет изменить правило расчета профиля нагрузки
categories: [Best practices]
tags: [best_practices, rps]
weight: 2
---

Обычно в тестах, когда мы увеличиваем скорость запросов, подаваемых на тестируемый севис, указывая схему `line`, `const`, `step` в секции rps,
то в секции `startup` мы указываем схему `once`, т.к. хотим, чтобы с самого начала теста нам были доступны все инстансы для того, чтобы сгенерить нужную нам нагрузку.

В тестах со сценарной нагрузкой, когда у нас каждый инстанс описыает поведение пользователя, то в секции `startup` можно указывать плавный рост количества пользователей, например схемой `instance_step`, увеличивая ступенчато их кол-во, или `const`, увеличивая пользователей с постоянной скоростью.
Для этого можно использовать настройку пула инстансов `rps-per-instance`. Она полезна для сценарного генератора, когда мы хотим ограничить скорость каждого пользователя в rps.

Например, укажем `const` и включим `rps-per-instance`, то потом увеличивая пользователей через `instance_step`, мы имитируем реальную пользовательскую нагрузку.

Пример:

```yaml
pools:
  - id: "scenario"
    ammo:
      type: http/scenario
      file: http_payload.hcl
    result:
      type: discard
    gun:
      target: localhost:443
      type: http/scenario
      answlog:
        enabled: false
    rps-per-instance: true
    rps:
      - type: const
        duration: 1m
        ops: 4
    startup:
      type: instance_step
      from: 10
      to: 100
      step: 10
      stepduration: 10s
log:
  level: error
```
