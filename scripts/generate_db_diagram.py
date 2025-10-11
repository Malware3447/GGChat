#!/usr/bin/env python3
"""
Скрипт для генерации Mermaid диаграммы базы данных из JSON схемы
"""

import json
import sys
import os

def generate_mermaid_diagram(schema_file):
    """Генерирует Mermaid диаграмму из JSON схемы БД"""
    
    try:
        with open(schema_file, 'r', encoding='utf-8') as f:
            schema = json.load(f)
    except FileNotFoundError:
        print(f"Файл {schema_file} не найден")
        return None
    except json.JSONDecodeError as e:
        print(f"Ошибка парсинга JSON: {e}")
        return None
    
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
        
        # Добавляем пустую строку между таблицами
        mermaid.append("")
    
    return "\n".join(mermaid)

def main():
    schema_file = "db_schema.json"
    
    if len(sys.argv) > 1:
        schema_file = sys.argv[1]
    
    if not os.path.exists(schema_file):
        print(f"Файл схемы {schema_file} не найден")
        sys.exit(1)
    
    diagram = generate_mermaid_diagram(schema_file)
    
    if diagram:
        print("Сгенерированная Mermaid диаграмма:")
        print("```mermaid")
        print(diagram)
        print("```")
        
        # Сохраняем в файл
        output_file = "database_diagram.mmd"
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(diagram)
        print(f"\nДиаграмма сохранена в файл: {output_file}")
    else:
        sys.exit(1)

if __name__ == "__main__":
    main()

