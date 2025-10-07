#!/usr/bin/env python3
"""
Скрипт для генерации полной визуализации БД с данными
"""

import json
import subprocess
import sys
import os
from datetime import datetime

def get_db_data():
    """Получает данные из PostgreSQL"""
    try:
        # Получаем список таблиц
        result = subprocess.run([
            'docker', 'exec', 'ggchat-postgres', 'psql', '-U', 'demo', '-d', 'ggchat', 
            '-t', '-c', "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"
        ], capture_output=True, text=True, check=True)
        
        tables = [line.strip() for line in result.stdout.split('\n') if line.strip()]
        
        data = {}
        for table in tables:
            # Получаем данные из каждой таблицы
            result = subprocess.run([
                'docker', 'exec', 'ggchat-postgres', 'psql', '-U', 'demo', '-d', 'ggchat',
                '-t', '-c', f"SELECT * FROM {table} LIMIT 10;"
            ], capture_output=True, text=True, check=True)
            
            # Получаем структуру таблицы
            structure_result = subprocess.run([
                'docker', 'exec', 'ggchat-postgres', 'psql', '-U', 'demo', '-d', 'ggchat',
                '-t', '-c', f"SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '{table}' ORDER BY ordinal_position;"
            ], capture_output=True, text=True, check=True)
            
            data[table] = {
                'structure': structure_result.stdout.strip(),
                'data': result.stdout.strip()
            }
        
        return data
    except subprocess.CalledProcessError as e:
        print(f"Ошибка подключения к БД: {e}")
        return None

def generate_enhanced_mermaid_diagram(schema_file):
    """Генерирует улучшенную Mermaid диаграмму с данными"""
    
    try:
        with open(schema_file, 'r', encoding='utf-8') as f:
            schema = json.load(f)
    except FileNotFoundError:
        print(f"Файл {schema_file} не найден")
        return None
    except json.JSONDecodeError as e:
        print(f"Ошибка парсинга JSON: {e}")
        return None
    
    # Получаем данные из БД
    db_data = get_db_data()
    
    mermaid = ["erDiagram"]
    
    for table in schema.get('tables', []):
        table_name = table['name']
        
        # Определяем столбцы таблицы
        columns = []
        for column in table.get('columns', []):
            col_name = column['name']
            col_type = column['type']
            
            if column.get('primary_key'):
                columns.append(f"{col_type} {col_name} PK")
            elif not column.get('nullable', True):
                columns.append(f"{col_type} {col_name} \"NOT NULL\"")
            else:
                columns.append(f"{col_type} {col_name}")
        
        # Добавляем таблицу в диаграмму
        mermaid.append(f"    {table_name} {{")
        for column in columns:
            mermaid.append(f"        {column}")
        mermaid.append("    }")
        
        # Добавляем связи
        for rel in table.get('relationships', []):
            if rel['type'] == 'one_to_many':
                mermaid.append(f"    {table_name} ||--o{{ {rel['target_table']} : \"{rel['foreign_key']}\"")
            elif rel['type'] == 'many_to_one':
                mermaid.append(f"    {table_name} }}o--|| {rel['target_table']} : \"{rel['foreign_key']}\"")
    
    return "\n".join(mermaid), db_data

def generate_data_summary_html(db_data):
    """Генерирует HTML с данными из БД"""
    if not db_data:
        return "<p>Нет данных для отображения</p>"
    
    html = ["<div class='database-data'>"]
    html.append("<h2>📊 Данные в базе данных</h2>")
    
    for table_name, table_info in db_data.items():
        html.append(f"<div class='table-section'>")
        html.append(f"<h3>🗃️ Таблица: {table_name}</h3>")
        
        # Структура таблицы
        html.append("<h4>Структура:</h4>")
        html.append("<pre>")
        html.append(table_info['structure'])
        html.append("</pre>")
        
        # Данные таблицы
        html.append("<h4>Данные (первые 10 записей):</h4>")
        if table_info['data']:
            html.append("<pre>")
            html.append(table_info['data'])
            html.append("</pre>")
        else:
            html.append("<p><em>Таблица пуста</em></p>")
        
        html.append("</div>")
    
    html.append("</div>")
    return "\n".join(html)

def main():
    schema_file = "db_schema.json"
    
    if len(sys.argv) > 1:
        schema_file = sys.argv[1]
    
    if not os.path.exists(schema_file):
        print(f"Файл схемы {schema_file} не найден")
        sys.exit(1)
    
    diagram, db_data = generate_enhanced_mermaid_diagram(schema_file)
    
    if diagram:
        # Создаем полную документацию с данными
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        
        full_doc = f"""# 🗄️ База данных GGChat - Полная визуализация

*Обновлено: {timestamp}*

## 📋 Диаграмма схемы базы данных

```mermaid
{diagram}
```

## 📊 Статистика таблиц

| Таблица | Описание |
|---------|----------|
| users | Пользователи системы |
| chats | Чаты/беседы |
| chat_nembers | Участники чатов |
| message | Сообщения |
| message_status | Статусы сообщений |

## 🔗 Связи между таблицами

- **users** → **chat_nembers** (один ко многим)
- **users** → **message** (один ко многим) 
- **users** → **message_status** (один ко многим)
- **chats** → **chat_nembers** (один ко многим)
- **chats** → **message** (один ко многим)
- **message** → **message_status** (один ко многим)

{generate_data_summary_html(db_data)}

## 🛠️ Команды для работы с БД

```bash
# Просмотр всех таблиц
task db:connect

# Генерация новой диаграммы
task db:schema

# Полная визуализация (этот файл)
python3 scripts/generate_db_dashboard.py

# Сброс БД
task db:reset
```
"""
        
        # Сохраняем в файл
        output_file = "database_full_dashboard.md"
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(full_doc)
        
        # Сохраняем отдельную диаграмму
        diagram_file = "database_diagram_enhanced.mmd"
        with open(diagram_file, 'w', encoding='utf-8') as f:
            f.write(diagram)
        
        print(f"✅ Полная визуализация создана: {output_file}")
        print(f"✅ Улучшенная диаграмма сохранена: {diagram_file}")
        print(f"📊 Данные из {len(db_data) if db_data else 0} таблиц включены")
        
    else:
        sys.exit(1)

if __name__ == "__main__":
    main()
