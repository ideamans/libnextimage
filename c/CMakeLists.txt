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
# imagedec/imageenc ライブラリを使用するため ANIM_UTILS を ON にする
# WebPAnimEncoder を使用するため WEBP_BUILD_LIBWEBPMUX を ON にする
set(WEBP_BUILD_ANIM_UTILS ON CACHE BOOL "" FORCE)
set(WEBP_BUILD_CWEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_DWEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_GIF2WEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_IMG2WEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_VWEBP OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_WEBPINFO OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_WEBPMUX OFF CACHE BOOL "" FORCE)
set(WEBP_BUILD_EXTRAS OFF CACHE BOOL "" FORCE)
set(WEBP_ENABLE_SIMD ON CACHE BOOL "" FORCE)
set(WEBP_BUILD_LIBWEBPMUX ON CACHE BOOL "" FORCE)
# インストールを無効化（libnextimage.aに統合するため）
set(CMAKE_SKIP_INSTALL_ALL_DEPENDENCY ON CACHE BOOL "" FORCE)
add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp ${CMAKE_CURRENT_BINARY_DIR}/libwebp EXCLUDE_FROM_ALL)

# libavif のビルド
# AOMコーデックのみを使用（シンプルな構成）
set(AVIF_CODEC_AOM LOCAL CACHE STRING "" FORCE)
set(AVIF_CODEC_DAV1D OFF CACHE BOOL "" FORCE)
set(AVIF_LIBYUV OFF CACHE BOOL "" FORCE)
set(AVIF_BUILD_APPS OFF CACHE BOOL "" FORCE)
set(AVIF_BUILD_TESTS OFF CACHE BOOL "" FORCE)
set(AVIF_BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
set(AVIF_ENABLE_WERROR OFF CACHE BOOL "" FORCE)
set(BUILD_SHARED_LIBS OFF CACHE BOOL "" FORCE)
add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../deps/libavif ${CMAKE_CURRENT_BINARY_DIR}/libavif EXCLUDE_FROM_ALL)

# インクルードディレクトリ
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/src
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/imageio
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/examples
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libavif/include
)

# ソースファイル
set(SOURCES
    src/common.c
    src/webp.c
    src/avif.c
    ../deps/libwebp/examples/gifdec.c
)

# 静的ライブラリのビルド
add_library(nextimage STATIC ${SOURCES})

# nextimageターゲットのインクルードディレクトリを追加
target_include_directories(nextimage PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/imageio
)

# GIF support を有効化（gifdec.c で使用）
target_compile_definitions(nextimage PRIVATE WEBP_HAVE_GIF)

# giflib のパスを探す
# NO_DEFAULT_PATH を使用して、明示的に指定したパスのみを検索（Mono framework などを除外）
find_library(GIF_LIBRARY NAMES gif
    PATHS /opt/homebrew/lib /usr/local/lib /usr/lib
    NO_DEFAULT_PATH)
find_path(GIF_INCLUDE_DIR gif_lib.h
    PATHS /opt/homebrew/include /usr/local/include /usr/include
    NO_DEFAULT_PATH)

if(GIF_LIBRARY AND GIF_INCLUDE_DIR)
    message(STATUS "Found giflib: ${GIF_LIBRARY}")
    message(STATUS "GIF include dir: ${GIF_INCLUDE_DIR}")
    # BEFORE を使用して、このパスを最優先にする（Mono framework より優先）
    target_include_directories(nextimage BEFORE PRIVATE ${GIF_INCLUDE_DIR})
else()
    message(WARNING "giflib not found - WebP to GIF conversion will not be available")
endif()

# libwebp と libavif のリンク - Goからはこのライブラリのみをリンクすれば良いように
# webp, webpdemux, webpmux, sharpyuv, imagedec, imageenc, avif を内部依存として持たせる
target_link_libraries(nextimage PUBLIC webp webpdemux libwebpmux sharpyuv imagedec imageenc avif)

# giflib をリンク（見つかった場合）
if(GIF_LIBRARY)
    target_link_libraries(nextimage PUBLIC ${GIF_LIBRARY})
endif()

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

    # Use libtool to combine libraries on macOS/BSD, or ar MRI script on Linux/Windows
    if(APPLE)
        # macOS: use libtool
        execute_process(
            COMMAND libtool -static -o \"\${LIB_DIR}/libnextimage.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libnextimage.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebp.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpdemux.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpmux.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libsharpyuv.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimagedec.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageenc.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageioutil.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/libavif/libavif_internal.a\"
                \"${CMAKE_CURRENT_BINARY_DIR}/_deps/aom-build/libaom.a\"
        )
    else()
        # Linux/Windows (MSYS2): use ar MRI script
        set(MRI_SCRIPT \"\${CMAKE_CURRENT_BINARY_DIR}/combine.mri\")
        file(WRITE \"\${MRI_SCRIPT}\" \"CREATE \${LIB_DIR}/libnextimage.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libnextimage.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebp.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpdemux.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpmux.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libsharpyuv.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimagedec.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageenc.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageioutil.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/libavif/libavif_internal.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB ${CMAKE_CURRENT_BINARY_DIR}/_deps/aom-build/libaom.a\\n\")
        file(APPEND \"\${MRI_SCRIPT}\" \"SAVE\\nEND\\n\")
        execute_process(
            COMMAND ar -M
            INPUT_FILE \"\${MRI_SCRIPT}\"
        )
        execute_process(
            COMMAND ranlib \"\${LIB_DIR}/libnextimage.a\"
        )
    endif()

    # Cleanup
    file(REMOVE_RECURSE \"\${TEMP_DIR}\")

    message(STATUS \"Combined library installed to: \${LIB_DIR}/libnextimage.a\")
")

install(FILES
    include/nextimage.h
    include/webp.h
    include/avif.h
    DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/../include
)

# テスト有効化
enable_testing()

# 基本テストプログラム
add_executable(basic_test test/basic_test.c)
target_link_libraries(basic_test nextimage)

# 新しいAPIのシンプルテスト
add_executable(simple_test test/simple_test.c)
target_link_libraries(simple_test nextimage)

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

# シンプルテスト（新API）
add_test(NAME SimpleTest COMMAND simple_test)

if(ENABLE_ASAN)
    target_compile_options(simple_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
    target_link_options(simple_test PRIVATE -fsanitize=address)
endif()

if(ENABLE_UBSAN)
    target_compile_options(simple_test PRIVATE -fsanitize=undefined)
    target_link_options(simple_test PRIVATE -fsanitize=undefined)
endif()

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

# AVIFテスト
add_executable(avif_test test/avif_test.c)
target_link_libraries(avif_test nextimage)

if(ENABLE_ASAN)
    target_compile_options(avif_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
    target_link_options(avif_test PRIVATE -fsanitize=address)
endif()

if(ENABLE_UBSAN)
    target_compile_options(avif_test PRIVATE -fsanitize=undefined)
    target_link_options(avif_test PRIVATE -fsanitize=undefined)
endif()

add_test(NAME AVIFTest COMMAND avif_test)

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

# ヘッダーコンパイルテスト（新しいインターフェース）
add_executable(header_test test/header_test.c)
target_link_libraries(header_test nextimage)
add_test(NAME HeaderTest COMMAND header_test)

# コマンドインターフェーステスト（SPEC.md準拠）
add_executable(command_interface_test test/command_interface_test.c)
target_link_libraries(command_interface_test nextimage)
add_test(NAME CommandInterfaceTest COMMAND command_interface_test)

# デコーダーテスト（dwebp/avifdec）
add_executable(decoder_test test/decoder_test.c)
target_link_libraries(decoder_test nextimage)
add_test(NAME DecoderTest COMMAND decoder_test)
