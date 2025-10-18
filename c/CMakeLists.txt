cmake_minimum_required(VERSION 3.15)
project(libnextimage C)

# C11標準を要求
set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# ビルドタイプの設定
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# デバッグビルドの定義
if(CMAKE_BUILD_TYPE MATCHES Debug)
    add_definitions(-DNEXTIMAGE_DEBUG)
endif()

# プラットフォーム検出
if(APPLE)
    if(CMAKE_SYSTEM_PROCESSOR MATCHES "arm64")
        set(PLATFORM "darwin-arm64")
    else()
        set(PLATFORM "darwin-amd64")
    endif()
elseif(UNIX)
    if(CMAKE_SYSTEM_PROCESSOR MATCHES "aarch64")
        set(PLATFORM "linux-arm64")
    else()
        set(PLATFORM "linux-amd64")
    endif()
elseif(WIN32)
    set(PLATFORM "windows-amd64")
else()
    set(PLATFORM "other")
endif()

message(STATUS "Platform: ${PLATFORM}")
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")

# libwebp のビルド
set(WEBP_BUILD_ANIM_UTILS OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_CWEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_DWEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_GIF2WEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_IMG2WEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_VWEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_WEBPINFO OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_WEBPMUX OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_EXTRAS OFF CACHE BOOL "" FORCE)
set(WEBP_ENABLE_SIMD ON CACHE BOOL "" FORCE)
set(WEBP_BUILD_LIBWEBPMUX OFF CACHE BOOL "" FORCE)
# インストールを無効化（libnextimage.aに統合するため）
set(CMAKE_SKIP_INSTALL_ALL_DEPENDENCY ON CACHE BOOL "" FORCE)
add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp ${CMAKE_CURRENT_BINARY_DIR}/libwebp EXCLUDE_FROM_ALL)

# インクルードディレクトリ
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/src
)

# ソースファイル
set(SOURCES
    src/common.c
    src/webp.c
)

# 静的ライブラリのビルド
add_library(nextimage STATIC ${SOURCES})

# libwebp のリンク - Goからはこのライブラリのみをリンクすれば良いように
# webp, webpdemux, sharpyuvを内部依存として持たせる
target_link_libraries(nextimage PUBLIC webp webpdemux sharpyuv)

# コンパイラ警告の有効化
if(CMAKE_C_COMPILER_ID MATCHES "GNU|Clang")
    target_compile_options(nextimage PRIVATE
        -Wall
        -Wextra
        -Wpedantic
        -Werror
    )
elseif(MSVC)
    target_compile_options(nextimage PRIVATE
        /W4
        /WX
    )
endif()

# Sanitizer オプション
option(ENABLE_ASAN "Enable AddressSanitizer" OFF)
option(ENABLE_UBSAN "Enable UndefinedBehaviorSanitizer" OFF)

if(ENABLE_ASAN)
    message(STATUS "AddressSanitizer enabled")
    target_compile_options(nextimage PRIVATE -fsanitize=address -fno-omit-frame-pointer)
    target_link_options(nextimage PRIVATE -fsanitize=address)
endif()

if(ENABLE_UBSAN)
    message(STATUS "UndefinedBehaviorSanitizer enabled")
    target_compile_options(nextimage PRIVATE -fsanitize=undefined)
    target_link_options(nextimage PRIVATE -fsanitize=undefined)
endif()

# インストール設定
# 複数の静的ライブラリを結合してlibnextimage.aに統合
install(CODE "
    message(STATUS \"Combining static libraries into libnextimage.a...\")

    set(LIB_DIR \"${CMAKE_CURRENT_SOURCE_DIR}/../lib/${PLATFORM}\")
    file(MAKE_DIRECTORY \"\${LIB_DIR}\")

    # Temporary directory for extraction
    set(TEMP_DIR \"\${CMAKE_CURRENT_BINARY_DIR}/temp_combine\")
    file(REMOVE_RECURSE \"\${TEMP_DIR}\")
    file(MAKE_DIRECTORY \"\${TEMP_DIR}\")

    # Extract all object files from each library
    execute_process(
        COMMAND ar -x \"${CMAKE_CURRENT_BINARY_DIR}/libnextimage.a\"
        WORKING_DIRECTORY \"\${TEMP_DIR}\"
    )
    execute_process(
        COMMAND ar -x \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebp.a\"
        WORKING_DIRECTORY \"\${TEMP_DIR}\"
    )
    execute_process(
        COMMAND ar -x \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpdemux.a\"
        WORKING_DIRECTORY \"\${TEMP_DIR}\"
    )
    execute_process(
        COMMAND ar -x \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libsharpyuv.a\"
        WORKING_DIRECTORY \"\${TEMP_DIR}\"
    )

    # Get all object files
    file(GLOB OBJ_FILES \"\${TEMP_DIR}/*.o\")

    # Create combined library
    execute_process(
        COMMAND ar -qc \"\${LIB_DIR}/libnextimage.a\" \${OBJ_FILES}
        WORKING_DIRECTORY \"\${TEMP_DIR}\"
    )

    # Run ranlib
    execute_process(
        COMMAND ranlib \"\${LIB_DIR}/libnextimage.a\"
    )

    # Cleanup
    file(REMOVE_RECURSE \"\${TEMP_DIR}\")

    message(STATUS \"Combined library installed to: \${LIB_DIR}/libnextimage.a\")
")

install(FILES
    include/nextimage.h
    include/webp.h
    DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/include
)

# テスト有効化
enable_testing()

# 基本テストプログラム
add_executable(basic_test test/basic_test.c)
target_link_libraries(basic_test nextimage)

# Sanitizerフラグをテストにも適用
if(ENABLE_ASAN)
    target_compile_options(basic_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
    target_link_options(basic_test PRIVATE -fsanitize=address)
endif()

if(ENABLE_UBSAN)
    target_compile_options(basic_test PRIVATE -fsanitize=undefined)
    target_link_options(basic_test PRIVATE -fsanitize=undefined)
endif()

add_test(NAME BasicTest COMMAND basic_test)

# WebPテスト
add_executable(webp_test test/webp_test.c)
target_link_libraries(webp_test nextimage)

if(ENABLE_ASAN)
    target_compile_options(webp_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
    target_link_options(webp_test PRIVATE -fsanitize=address)
endif()

if(ENABLE_UBSAN)
    target_compile_options(webp_test PRIVATE -fsanitize=undefined)
    target_link_options(webp_test PRIVATE -fsanitize=undefined)
endif()

add_test(NAME WebPTest COMMAND webp_test)

# デバッグビルドのみ: リークカウンターテスト
if(CMAKE_BUILD_TYPE MATCHES Debug)
    add_executable(leak_counter_test test/leak_counter_test.c)
    target_link_libraries(leak_counter_test nextimage)
    add_test(NAME LeakCounterTest COMMAND leak_counter_test)
endif()

# Sanitizerビルドのみ: Sanitizer専用テスト
if(ENABLE_ASAN OR ENABLE_UBSAN)
    add_executable(sanitizer_test test/sanitizer/sanitizer_test.c)
    target_link_libraries(sanitizer_test nextimage)

    if(ENABLE_ASAN)
        target_compile_options(sanitizer_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
        target_link_options(sanitizer_test PRIVATE -fsanitize=address)
    endif()

    if(ENABLE_UBSAN)
        target_compile_options(sanitizer_test PRIVATE -fsanitize=undefined)
        target_link_options(sanitizer_test PRIVATE -fsanitize=undefined)
    endif()

    add_test(NAME SanitizerTest COMMAND sanitizer_test)
endif()
