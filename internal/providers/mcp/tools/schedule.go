package tools

const scheduleAddSchema = `
{
  "name": "schedule_add",
  "description": "Запланировать выполнение новой задачи (напоминание или регулярное действие).",
  "parameters": {
    "type": "object",
    "properties": {
      "type": {
        "type": "string",
        "enum": ["one_off", "interval"],
        "description": "Тип задачи: 'one_off' (один раз) или 'interval' (повторять)."
      },
      "time_spec": {
        "type": "string",
        "description": "Для 'one_off': дата ISO8601 (2023-10-27T15:00:00Z) или задержка (10m, 2h). Для 'interval': период повторения (30s, 10m, 24h)."
      },
      "task_name": {
        "type": "string",
        "description": "Уникальное короткое название задачи (slug), без пробелов (напр: 'daily-digest', 'remind-milk')."
      },
      "instruction": {
        "type": "string",
        "description": "Подробная инструкция для агента, который проснется в будущем. Что именно нужно сделать?"
      }
    },
    "required": ["type", "time_spec", "task_name", "instruction"]
  }
}
`

const scheduleListSchema = `
{
  "name": "schedule_list",
  "description": "Получить список всех активных запланированных задач.",
  "parameters": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
`

const scheduleCancelSchema = `
{
  "name": "schedule_cancel",
  "description": "Отменить и удалить запланированную задачу по её названию.",
  "parameters": {
    "type": "object",
    "properties": {
      "task_id": {
        "type": "string",
        "description": "Название задачи (task_name), которую нужно удалить."
      }
    },
    "required": ["task_id"]
  }
}
`
