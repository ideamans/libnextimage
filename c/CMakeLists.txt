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

# システムライブラリの検索（静的ライブラリを優先）

if(APPLE)
    # macOS: Use NO_DEFAULT_PATH to avoid Mono framework conflicts
    # jpeg
    find_library(JPEG_LIBRARY NAMES libjpeg.a jpeg
        PATHS /opt/homebrew/lib /usr/local/lib /usr/lib
        NO_DEFAULT_PATH)
    find_path(JPEG_INCLUDE_DIR jpeglib.h
        PATHS /opt/homebrew/include /usr/local/include /usr/include
        NO_DEFAULT_PATH)

    # png
    find_library(PNG_LIBRARY NAMES libpng.a png
        PATHS /opt/homebrew/lib /usr/local/lib /usr/lib
        NO_DEFAULT_PATH)
    find_path(PNG_INCLUDE_DIR png.h
        PATHS /opt/homebrew/include /usr/local/include /usr/include
        NO_DEFAULT_PATH)
else()
    # Linux/Windows: Use standard CMake find modules (includes multi-arch paths)
    # Add architecture-specific library paths for Linux
    if(UNIX AND NOT APPLE)
        # Get the architecture-specific library directory
        execute_process(
            COMMAND gcc -print-multiarch
            OUTPUT_VARIABLE GCC_MULTIARCH
            OUTPUT_STRIP_TRAILING_WHITESPACE
            ERROR_QUIET
        )
        if(GCC_MULTIARCH)
            set(MULTIARCH_LIB_PATH "/usr/lib/${GCC_MULTIARCH}")
            message(STATUS "Detected multiarch path: ${MULTIARCH_LIB_PATH}")
        else()
            message(STATUS "gcc -print-multiarch returned empty, using default paths")
        endif()
    endif()

    # jpeg
    if(MULTIARCH_LIB_PATH)
        find_library(JPEG_LIBRARY NAMES libjpeg.a jpeg
            PATHS ${MULTIARCH_LIB_PATH} /usr/lib /usr/local/lib)
    else()
        find_library(JPEG_LIBRARY NAMES libjpeg.a jpeg)
    endif()
    find_path(JPEG_INCLUDE_DIR jpeglib.h)
    message(STATUS "JPEG library: ${JPEG_LIBRARY}")

    # png
    if(MULTIARCH_LIB_PATH)
        find_library(PNG_LIBRARY NAMES libpng.a png libpng16.a png16
            PATHS ${MULTIARCH_LIB_PATH} /usr/lib /usr/local/lib)
    else()
        find_library(PNG_LIBRARY NAMES libpng.a png libpng16.a png16)
    endif()
    find_path(PNG_INCLUDE_DIR png.h)
    message(STATUS "PNG library: ${PNG_LIBRARY}")
endif()

# zlib (macOS SDK or system library)
if(APPLE)
    # macOS: zlib is in the SDK
    execute_process(
        COMMAND xcrun --show-sdk-path
        OUTPUT_VARIABLE MACOS_SDK_PATH
        OUTPUT_STRIP_TRAILING_WHITESPACE
    )
    find_library(Z_LIBRARY NAMES z
        PATHS ${MACOS_SDK_PATH}/usr/lib /usr/lib
        NO_DEFAULT_PATH)
    find_path(Z_INCLUDE_DIR zlib.h
        PATHS ${MACOS_SDK_PATH}/usr/include /usr/include
        NO_DEFAULT_PATH)
else()
    # Linux/Windows: Use standard CMake find modules
    if(MULTIARCH_LIB_PATH)
        find_library(Z_LIBRARY NAMES libz.a z
            PATHS ${MULTIARCH_LIB_PATH} /usr/lib /usr/local/lib)
    else()
        find_library(Z_LIBRARY NAMES libz.a z)
    endif()
    find_path(Z_INCLUDE_DIR zlib.h)
    message(STATUS "zlib library: ${Z_LIBRARY}")
endif()

# giflib
if(APPLE)
    # macOS: explicit paths
    find_library(GIF_LIBRARY NAMES libgif.a gif
        PATHS /opt/homebrew/lib /usr/local/lib /usr/lib
        NO_DEFAULT_PATH)
    find_path(GIF_INCLUDE_DIR gif_lib.h
        PATHS /opt/homebrew/include /usr/local/include /usr/include
        NO_DEFAULT_PATH)
else()
    # Linux/Windows: Use standard CMake find modules
    if(MULTIARCH_LIB_PATH)
        find_library(GIF_LIBRARY NAMES libgif.a gif
            PATHS ${MULTIARCH_LIB_PATH} /usr/lib /usr/local/lib)
    else()
        find_library(GIF_LIBRARY NAMES libgif.a gif)
    endif()
    find_path(GIF_INCLUDE_DIR gif_lib.h)
    message(STATUS "giflib library: ${GIF_LIBRARY}")
endif()

# 検出結果を表示
if(JPEG_LIBRARY AND JPEG_INCLUDE_DIR)
    message(STATUS "Found libjpeg: ${JPEG_LIBRARY}")
    message(STATUS "  Include dir: ${JPEG_INCLUDE_DIR}")
else()
    message(FATAL_ERROR "libjpeg not found. Searched in standard CMake paths. Please install libjpeg-dev (Debian/Ubuntu) or jpeg (macOS).")
endif()

if(PNG_LIBRARY AND PNG_INCLUDE_DIR)
    message(STATUS "Found libpng: ${PNG_LIBRARY}")
    message(STATUS "  Include dir: ${PNG_INCLUDE_DIR}")
else()
    message(FATAL_ERROR "libpng not found. Please install libpng-dev (Debian/Ubuntu) or libpng (macOS).")
endif()

if(Z_LIBRARY AND Z_INCLUDE_DIR)
    message(STATUS "Found zlib: ${Z_LIBRARY}")
    message(STATUS "  Include dir: ${Z_INCLUDE_DIR}")
else()
    message(FATAL_ERROR "zlib not found. Please install zlib1g-dev (Debian/Ubuntu) or use system zlib (macOS).")
endif()

if(GIF_LIBRARY AND GIF_INCLUDE_DIR)
    message(STATUS "Found giflib: ${GIF_LIBRARY}")
    message(STATUS "GIF include dir: ${GIF_INCLUDE_DIR}")
    # libwebp が giflib を使用する際に、正しいパスを参照するように設定
    include_directories(BEFORE SYSTEM ${GIF_INCLUDE_DIR})
else()
    message(WARNING "giflib not found - WebP to GIF conversion will not be available")
endif()

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
)

# giflibが見つかった場合のみgifdec.cを追加
if(GIF_LIBRARY AND GIF_INCLUDE_DIR)
    list(APPEND SOURCES ../deps/libwebp/examples/gifdec.c)
endif()

# 静的ライブラリのビルド
add_library(nextimage STATIC ${SOURCES})

# 動的ライブラリのビルド
add_library(nextimage_shared SHARED ${SOURCES})

# 動的ライブラリの出力名を設定（全プラットフォーム共通）
set_target_properties(nextimage_shared PROPERTIES
    OUTPUT_NAME "nextimage"
)

# macOS: 動的ライブラリのインストール名を設定
if(APPLE)
    set_target_properties(nextimage_shared PROPERTIES
        INSTALL_NAME_DIR "@rpath"
        BUILD_WITH_INSTALL_RPATH TRUE
    )
endif()

# nextimageターゲットのインクルードディレクトリを追加
target_include_directories(nextimage PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/imageio
)

# nextimage_shared ターゲットのインクルードディレクトリを追加
target_include_directories(nextimage_shared PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/../deps/libwebp/imageio
)

# GIF support を有効化（giflibが見つかった場合のみ）
if(GIF_LIBRARY AND GIF_INCLUDE_DIR)
    target_compile_definitions(nextimage PRIVATE WEBP_HAVE_GIF)
    target_compile_definitions(nextimage_shared PRIVATE WEBP_HAVE_GIF)
endif()

# giflib のインクルードパスを nextimage ターゲットにも明示的に追加
if(GIF_INCLUDE_DIR)
    target_include_directories(nextimage BEFORE PRIVATE ${GIF_INCLUDE_DIR})
    target_include_directories(nextimage_shared BEFORE PRIVATE ${GIF_INCLUDE_DIR})
endif()

# libwebp と libavif のリンク - Goからはこのライブラリのみをリンクすれば良いように
# webp, webpdemux, webpmux, sharpyuv, imagedec, imageenc, avif を内部依存として持たせる
target_link_libraries(nextimage PUBLIC webp webpdemux libwebpmux sharpyuv imagedec imageenc avif)
target_link_libraries(nextimage_shared PUBLIC webp webpdemux libwebpmux sharpyuv imagedec imageenc avif)

# システムライブラリを静的リンク（libnextimage.aに含める）
target_link_libraries(nextimage PUBLIC ${JPEG_LIBRARY} ${PNG_LIBRARY} ${Z_LIBRARY})

# 動的ライブラリのシステムライブラリリンク
# Linux: 共有ライブラリには動的ライブラリをリンク（-fPIC問題回避）
# macOS: 静的ライブラリで問題なし
if(UNIX AND NOT APPLE)
    # Linux: 動的ライブラリを使用
    target_link_libraries(nextimage_shared PRIVATE jpeg png z)
else()
    # macOS/Windows: 静的ライブラリを使用
    target_link_libraries(nextimage_shared PRIVATE ${JPEG_LIBRARY} ${PNG_LIBRARY} ${Z_LIBRARY})
endif()

# macOS/Linux: リンカーに静的ライブラリを強制的に埋め込む
if(UNIX)
    # -Wl,--whole-archive は Linux で依存ライブラリ全体を含める
    # macOS では -Wl,-all_load を使用
    if(APPLE)
        # macOS: 動的ライブラリに静的ライブラリの内容を埋め込む
        target_link_options(nextimage_shared PRIVATE -Wl,-all_load)
    else()
        # Linux: 動的ライブラリに静的ライブラリの内容を埋め込む
        target_link_options(nextimage_shared PRIVATE -Wl,--exclude-libs,ALL)
    endif()
endif()

# giflib をリンク（見つかった場合）
if(GIF_LIBRARY)
    target_link_libraries(nextimage PUBLIC ${GIF_LIBRARY})

    # 共有ライブラリ向け: Linux は動的、それ以外は静的
    if(UNIX AND NOT APPLE)
        target_link_libraries(nextimage_shared PRIVATE gif)
    else()
        target_link_libraries(nextimage_shared PRIVATE ${GIF_LIBRARY})
    endif()
endif()

# コンパイラ警告の有効化
if(CMAKE_C_COMPILER_ID MATCHES "GNU|Clang")
    target_compile_options(nextimage PRIVATE
        -Wall
        -Wextra
        -Wpedantic
        -Werror
    )
    target_compile_options(nextimage_shared PRIVATE
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
    target_compile_options(nextimage_shared PRIVATE
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

    set(STATIC_LIB_DIR \"${CMAKE_CURRENT_SOURCE_DIR}/../lib/static\")
    file(MAKE_DIRECTORY \"\${STATIC_LIB_DIR}\")

    # Collect required libraries
    set(LIBS_TO_COMBINE
        \"${CMAKE_CURRENT_BINARY_DIR}/libnextimage.a\"
        \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebp.a\"
        \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpdemux.a\"
        \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libwebpmux.a\"
        \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libsharpyuv.a\"
        \"${CMAKE_CURRENT_BINARY_DIR}/libavif/libavif_internal.a\"
        \"${CMAKE_CURRENT_BINARY_DIR}/_deps/aom-build/libaom.a\"
    )

    # Add optional libraries if they exist (imagedec/enc/ioutil may not exist without giflib)
    if(EXISTS \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimagedec.a\")
        list(APPEND LIBS_TO_COMBINE \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimagedec.a\")
        message(STATUS \"Including libimagedec.a\")
    else()
        message(STATUS \"Skipping libimagedec.a (not built)\")
    endif()
    if(EXISTS \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageenc.a\")
        list(APPEND LIBS_TO_COMBINE \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageenc.a\")
        message(STATUS \"Including libimageenc.a\")
    else()
        message(STATUS \"Skipping libimageenc.a (not built)\")
    endif()
    if(EXISTS \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageioutil.a\")
        list(APPEND LIBS_TO_COMBINE \"${CMAKE_CURRENT_BINARY_DIR}/libwebp/libimageioutil.a\")
        message(STATUS \"Including libimageioutil.a\")
    else()
        message(STATUS \"Skipping libimageioutil.a (not built)\")
    endif()

    # Add system libraries (jpeg, png, gif)
    # Note: zlib is excluded because on macOS it's a .tbd file (dynamic library stub)
    # and cannot be statically linked. It must be linked with -lz at Go build time.
    list(APPEND LIBS_TO_COMBINE \"${JPEG_LIBRARY}\")
    list(APPEND LIBS_TO_COMBINE \"${PNG_LIBRARY}\")

    # Add giflib if it was found during configuration
    # Use the GIF_LIBRARY variable that was already found earlier
    if(\"${GIF_LIBRARY}\" STREQUAL \"GIF_LIBRARY-NOTFOUND\" OR \"${GIF_LIBRARY}\" STREQUAL \"\")
        message(WARNING \"giflib not found - WebP to GIF conversion will not be available\")
    else()
        list(APPEND LIBS_TO_COMBINE \"${GIF_LIBRARY}\")
        message(STATUS \"Including giflib: ${GIF_LIBRARY}\")
    endif()

    # Platform-specific combining strategy
    if(APPLE)
        # macOS: use libtool -static (properly handles nested archives)
        execute_process(
            COMMAND libtool -static -o \"\${STATIC_LIB_DIR}/libnextimage.a\" \${LIBS_TO_COMBINE}
        )
    else()
        # Linux/Windows: use ar MRI script
        set(MRI_SCRIPT \"\${CMAKE_CURRENT_BINARY_DIR}/combine.mri\")
        file(WRITE \"\${MRI_SCRIPT}\" \"CREATE \${STATIC_LIB_DIR}/libnextimage.a\\n\")
        foreach(LIB \${LIBS_TO_COMBINE})
            file(APPEND \"\${MRI_SCRIPT}\" \"ADDLIB \${LIB}\\n\")
        endforeach()
        file(APPEND \"\${MRI_SCRIPT}\" \"SAVE\\nEND\\n\")
        execute_process(
            COMMAND ar -M
            INPUT_FILE \"\${MRI_SCRIPT}\"
        )
        execute_process(
            COMMAND ranlib \"\${STATIC_LIB_DIR}/libnextimage.a\"
        )
    endif()

    message(STATUS \"Combined library installed to: \${STATIC_LIB_DIR}/libnextimage.a\")
")

# 動的ライブラリのインストール
install(CODE "
    message(STATUS \"Installing shared library...\")

    set(SHARED_LIB_DIR \"${CMAKE_CURRENT_SOURCE_DIR}/../lib/shared\")
    file(MAKE_DIRECTORY \"\${SHARED_LIB_DIR}\")

    # Platform-specific shared library extension
    if(APPLE)
        set(SHARED_LIB_EXT \"dylib\")
    elseif(WIN32)
        set(SHARED_LIB_EXT \"dll\")
    else()
        set(SHARED_LIB_EXT \"so\")
    endif()

    # Copy shared library to shared directory
    file(COPY \"${CMAKE_CURRENT_BINARY_DIR}/libnextimage.\${SHARED_LIB_EXT}\"
         DESTINATION \"\${SHARED_LIB_DIR}\")

    message(STATUS \"Shared library installed to: \${SHARED_LIB_DIR}/libnextimage.\${SHARED_LIB_EXT}\")
")

# ヘッダファイルのインストール
install(FILES
    include/nextimage.h
    include/webp.h
    include/avif.h
    DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/../lib/include
)

# nextimage サブディレクトリのヘッダファイルもインストール
install(DIRECTORY
    include/nextimage
    DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/../lib/include
    FILES_MATCHING PATTERN "*.h"
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

# WebPテスト - 古いAPIを使用しているため一時的に無効化
# TODO: 新しいAPI（nextimage_webp_encode_alloc(input_data, input_size, options, output)）に対応
# add_executable(webp_test test/webp_test.c)
# target_link_libraries(webp_test nextimage)
#
# if(ENABLE_ASAN)
#     target_compile_options(webp_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
#     target_link_options(webp_test PRIVATE -fsanitize=address)
# endif()
#
# if(ENABLE_UBSAN)
#     target_compile_options(webp_test PRIVATE -fsanitize=undefined)
#     target_link_options(webp_test PRIVATE -fsanitize=undefined)
# endif()
#
# add_test(NAME WebPTest COMMAND webp_test)

# AVIFテスト - 古いAPIを使用しているため一時的に無効化
# TODO: 新しいAPI（nextimage_avif_encode_alloc(input_data, input_size, options, output)）に対応
# add_executable(avif_test test/avif_test.c)
# target_link_libraries(avif_test nextimage)
#
# if(ENABLE_ASAN)
#     target_compile_options(avif_test PRIVATE -fsanitize=address -fno-omit-frame-pointer)
#     target_link_options(avif_test PRIVATE -fsanitize=address)
# endif()
#
# if(ENABLE_UBSAN)
#     target_compile_options(avif_test PRIVATE -fsanitize=undefined)
#     target_link_options(avif_test PRIVATE -fsanitize=undefined)
# endif()
#
# add_test(NAME AVIFTest COMMAND avif_test)

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

# ========================================
# 動的ライブラリのテスト
# ========================================
# 動的ライブラリが正しくビルドされ、シンボルが正しくエクスポートされているか確認

# シンプルテスト（動的ライブラリ版）
add_executable(simple_test_shared test/simple_test.c)
target_link_libraries(simple_test_shared nextimage_shared)
add_test(NAME SimpleTestShared COMMAND simple_test_shared)

# WebPテスト（動的ライブラリ版） - 古いAPIを使用しているため一時的に無効化
# add_executable(webp_test_shared test/webp_test.c)
# target_link_libraries(webp_test_shared nextimage_shared)
# add_test(NAME WebPTestShared COMMAND webp_test_shared)

# AVIFテスト（動的ライブラリ版） - 古いAPIを使用しているため一時的に無効化
# add_executable(avif_test_shared test/avif_test.c)
# target_link_libraries(avif_test_shared nextimage_shared)
# add_test(NAME AVIFTestShared COMMAND avif_test_shared)
