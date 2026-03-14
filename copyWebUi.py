import os

# 配置你要打包的项目目录和输出文件名
PROJECT_DIR = "./web"
OUTPUT_FILE = "xingbeichuangshuo.txt"

# 配置你需要包含的文件后缀，排除图片、日志等无关文件
ALLOWED_EXTENSIONS = {'.vue', '.ts', '.js', '.html', '.css', '.json'}
# 排除不需要的文件夹（如依赖包、编译结果）
EXCLUDED_DIRS = {'.git', 'node_modules', 'venv', '__pycache__', 'build', 'dist', 'public'}

def pack_code():
    with open(OUTPUT_FILE, 'w', encoding='utf-8') as out_f:
        # 先生成一个项目目录树结构（可选，但有助于 AI 理解架构）
        out_f.write("=== 项目文件结构 ===\n")
        for root, dirs, files in os.walk(PROJECT_DIR):
            dirs[:] = [d for d in dirs if d not in EXCLUDED_DIRS]
            level = root.replace(PROJECT_DIR, '').count(os.sep)
            indent = ' ' * 4 * (level)
            out_f.write(f"{indent}{os.path.basename(root)}/\n")
            subindent = ' ' * 4 * (level + 1)
            for f in files:
                if os.path.splitext(f)[1] in ALLOWED_EXTENSIONS:
                    out_f.write(f"{subindent}{f}\n")

        out_f.write("\n\n=== 具体代码内容 ===\n")
        # 遍历写入文件内容
        for root, dirs, files in os.walk(PROJECT_DIR):
            dirs[:] = [d for d in dirs if d not in EXCLUDED_DIRS]
            for file in files:
                if os.path.splitext(file)[1] in ALLOWED_EXTENSIONS:
                    filepath = os.path.join(root, file)
                    out_f.write(f"\n\n--- 文件路径: {filepath} ---\n\n")
                    try:
                        with open(filepath, 'r', encoding='utf-8') as in_f:
                            out_f.write(in_f.read())
                    except Exception as e:
                        out_f.write(f"[读取失败: {e}]\n")

    print(f"打包完成！请将 {OUTPUT_FILE} 上传给 Gemini。")

if __name__ == "__main__":
    pack_code()