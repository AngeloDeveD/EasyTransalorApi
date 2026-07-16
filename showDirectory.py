import os
from pathlib import Path

# Список папок/файлов, которые мы хотим скрыть, чтобы не засорять вывод
IGNORED = {'.git', '__pycache__', 'node_modules', '.idea', '.vscode', 'venv'}

def build_tree(directory, prefix="", output_lines=None):
    if output_lines is None:
        output_lines = []

    # Получаем все элементы в текущей папке
    try:
        items = sorted(Path(directory).iterdir(), key=lambda x: (not x.is_dir(), x.name.lower()))
    except PermissionError:
        # Если нет прав доступа к папке — просто пропускаем
        return output_lines

    # Фильтруем скрытые и ненужные папки/файлы
    items = [item for item in items if item.name not in IGNORED and not item.name.startswith('.')]

    for i, item in enumerate(items):
        is_last = i == len(items) - 1
        
        # Выбираем правильный символ соединения
        connector = "└── " if is_last else "├── "
        
        # Если это папка, добавляем слэш в конце
        display_name = item.name + "\\" if item.is_dir() else item.name
        
        # Добавляем строку в наш список
        output_lines.append(f"{prefix}{connector}{display_name}")
        
        # Если это папка, заходим в неё рекурсивно
        if item.is_dir():
            # Для вложенных элементов меняем отступ: вертикальная линия или пробелы
            new_prefix = prefix + ("    " if is_last else "│   ")
            build_tree(item, new_prefix, output_lines)

    return output_lines

def save_directory_tree(root_dir, output_filename="directory_tree.txt"):
    root_path = Path(root_dir)
    
    # Начальная строка с названием корневой папки
    header = f"{root_path.name}\\"
    print(f"Сканирую папку: {root_path.resolve()}...")
    
    # Строим дерево
    tree_lines = build_tree(root_path, prefix="", output_lines=[header])
    
    # Сохраняем в файл
    with open(output_filename, "w", encoding="utf-8") as f:
        f.write("\n".join(tree_lines))
        
    print(f"Готово! Дерево сохранено в файл: {output_filename}")

# ================================
# ИСПОЛЬЗОВАНИЕ
# ================================
# Укажи здесь путь к папке, которую хочешь сохранить
target_folder = "."  # Точка означает текущую папку, можешь заменить на "C:/МояПапка"

save_directory_tree(target_folder)