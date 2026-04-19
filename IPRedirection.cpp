#include <filesystem>
#include <fstream>
#include <iostream>
#include <string>
#include <vector>
#ifdef _WIN32
#include <windows.h>
#endif

namespace fs = std::filesystem;

namespace {

constexpr std::streamoff kPrimaryOffset = 0x7D3EF40;
constexpr std::streamoff kSecondaryOffset = 0x7D54F8;

const std::string kExpectedPrimary = "https://kards.live.1939api.com/";
const std::string kExpectedSecondary = "https://kards.live.1939api.com/config";

bool fits_in_slot(const std::string& replacement, const std::string& original) {
    return replacement.size() <= original.size();
}

std::string trim(const std::string& text) {
    const auto begin = text.find_first_not_of(" \t\r\n");
    if (begin == std::string::npos) {
        return "";
    }

    const auto end = text.find_last_not_of(" \t\r\n");
    return text.substr(begin, end - begin + 1);
}

std::string strip_trailing_slashes(std::string value) {
    while (!value.empty() && (value.back() == '/' || value.back() == '\\')) {
        value.pop_back();
    }
    return value;
}

std::string normalize_base_url(std::string value) {
    value = trim(value);
    value = strip_trailing_slashes(value);
    return value;
}

bool has_http_scheme(const std::string& value) {
    return value.rfind("http://", 0) == 0 || value.rfind("https://", 0) == 0;
}

std::string build_primary_url(const std::string& base_url) {
    return base_url + "/";
}

std::string build_secondary_url(const std::string& base_url) {
    return base_url + "/config";
}

std::vector<char> read_file(const fs::path& path) {
    std::ifstream input(path, std::ios::binary);
    if (!input) {
        throw std::runtime_error("Failed to open the input file.");
    }

    input.seekg(0, std::ios::end);
    const std::streamsize size = input.tellg();
    input.seekg(0, std::ios::beg);

    if (size <= 0) {
        throw std::runtime_error("The input file is empty or its size could not be read.");
    }

    std::vector<char> buffer(static_cast<size_t>(size));
    if (!input.read(buffer.data(), size)) {
        throw std::runtime_error("Failed to read the input file.");
    }

    return buffer;
}

bool patch_at_offset(std::vector<char>& buffer,
                     std::streamoff offset,
                     const std::string& expected,
                     const std::string& replacement) {
    if (offset < 0) {
        return false;
    }

    const auto index = static_cast<size_t>(offset);
    if (index + expected.size() > buffer.size()) {
        return false;
    }

    const std::string current(buffer.data() + index, expected.size());
    if (current != expected) {
        return false;
    }

    std::fill(buffer.begin() + index, buffer.begin() + index + expected.size(), '\0');
    std::copy(replacement.begin(), replacement.end(), buffer.begin() + index);
    return true;
}

size_t patch_all_matches(std::vector<char>& buffer,
                         const std::string& expected,
                         const std::string& replacement) {
    if (expected.empty() || expected.size() > buffer.size()) {
        return 0;
    }

    size_t patched = 0;
    for (size_t i = 0; i + expected.size() <= buffer.size(); ++i) {
        bool matched = true;
        for (size_t j = 0; j < expected.size(); ++j) {
            if (buffer[i + j] != expected[j]) {
                matched = false;
                break;
            }
        }

        if (!matched) {
            continue;
        }

        std::fill(buffer.begin() + i, buffer.begin() + i + expected.size(), '\0');
        std::copy(replacement.begin(), replacement.end(), buffer.begin() + i);
        ++patched;
        i += expected.size() - 1;
    }

    return patched;
}

fs::path build_output_path(const fs::path& input_path) {
    return input_path.parent_path() / "kards-Win64-Shipping-Edited.exe";
}

void write_file(const fs::path& path, const std::vector<char>& buffer) {
    std::ofstream output(path, std::ios::binary);
    if (!output) {
        throw std::runtime_error("Failed to create the output file.");
    }

    output.write(buffer.data(), static_cast<std::streamsize>(buffer.size()));
    if (!output) {
        throw std::runtime_error("Failed to write the output file.");
    }
}

}  // namespace

int main() {
    try {
#ifdef _WIN32
        SetConsoleOutputCP(CP_UTF8);
        SetConsoleCP(CP_UTF8);
#endif

        std::cout << "请输入 kards-Win64-Shipping.exe 的完整路径：";
        std::string input_path_text;
        std::getline(std::cin, input_path_text);

        if (input_path_text.empty()) {
            std::cerr << "未输入文件路径。\n";
            return 1;
        }

        fs::path input_path = fs::path(input_path_text);
        if (!fs::exists(input_path)) {
            std::cerr << "文件不存在：" << input_path << '\n';
            return 1;
        }

        std::cout << "请输入新的服务器地址（必须包含 http:// 或 https://）：";
        std::string base_url_input;
        std::getline(std::cin, base_url_input);

        const std::string normalized_base_url = normalize_base_url(base_url_input);
        if (normalized_base_url.empty()) {
            std::cerr << "未输入服务器地址。\n";
            return 1;
        }

        if (!has_http_scheme(normalized_base_url)) {
            std::cerr << "服务器地址必须以 http:// 或 https:// 开头。\n";
            return 1;
        }

        const std::string replacement_primary = build_primary_url(normalized_base_url);
        const std::string replacement_secondary = build_secondary_url(normalized_base_url);

        if (!fits_in_slot(replacement_primary, kExpectedPrimary) ||
            !fits_in_slot(replacement_secondary, kExpectedSecondary)) {
            std::cerr << "替换后的地址过长，请使用更短的服务器地址。\n";
            return 1;
        }

        auto buffer = read_file(input_path);

        bool primary_offset_patched =
            patch_at_offset(buffer, kPrimaryOffset, kExpectedPrimary, replacement_primary);
        bool secondary_offset_patched =
            patch_at_offset(buffer, kSecondaryOffset, kExpectedSecondary, replacement_secondary);

        size_t primary_search_patched = 0;
        size_t secondary_search_patched = 0;

        if (!primary_offset_patched) {
            primary_search_patched = patch_all_matches(
                buffer, kExpectedPrimary, replacement_primary);
        }

        if (!secondary_offset_patched) {
            secondary_search_patched = patch_all_matches(
                buffer, kExpectedSecondary, replacement_secondary);
        }

        if (!primary_offset_patched && primary_search_patched == 0) {
            std::cerr << "未找到主服务器地址：" << kExpectedPrimary << '\n';
            return 1;
        }

        if (!secondary_offset_patched && secondary_search_patched == 0) {
            std::cerr << "未找到配置地址：" << kExpectedSecondary << '\n';
            return 1;
        }

        const fs::path output_path = build_output_path(input_path);
        write_file(output_path, buffer);

        std::cout << "修改完成。\n";
        std::cout << "输出文件：" << output_path << '\n';
        std::cout << "主地址替换为：" << replacement_primary << '\n';
        std::cout << "配置地址替换为：" << replacement_secondary << '\n';
        std::cout << "主地址替换结果："
                  << (primary_offset_patched ? "命中固定偏移"
                                             : "搜索命中 " + std::to_string(primary_search_patched) + " 处")
                  << '\n';
        std::cout << "配置地址替换结果："
                  << (secondary_offset_patched ? "命中固定偏移"
                                               : "搜索命中 " + std::to_string(secondary_search_patched) + " 处")
                  << '\n';
        return 0;
    } catch (const std::exception& ex) {
        std::cerr << "处理失败：" << ex.what() << '\n';
        return 1;
    }
}
