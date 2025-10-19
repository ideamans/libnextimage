#include "nextimage.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdarg.h>

// スレッドローカルストレージ用
#if defined(__GNUC__) || defined(__clang__)
    // GCC/Clang拡張（macOS含む）
    static __thread char g_error_buffer[1024] = {0};
#elif defined(_MSC_VER)
    // MSVC拡張
    static __declspec(thread) char g_error_buffer[1024] = {0};
#elif defined(__STDC_VERSION__) && __STDC_VERSION__ >= 201112L && !defined(__APPLE__)
    // C11 thread_local（macOS以外）
    #include <threads.h>
    static thread_local char g_error_buffer[1024] = {0};
#else
    // フォールバック: スレッドセーフではない
    #warning "Thread-local storage not supported, error messages may not be thread-safe"
    static char g_error_buffer[1024] = {0};
#endif

// デバッグビルド専用: メモリリークカウンター
#ifdef NEXTIMAGE_DEBUG
#include <stdatomic.h>
static atomic_int_least64_t g_allocation_counter = 0;

void nextimage_increment_alloc_counter(void) {
    atomic_fetch_add(&g_allocation_counter, 1);
}

void nextimage_decrement_alloc_counter(void) {
    atomic_fetch_sub(&g_allocation_counter, 1);
}

int64_t nextimage_allocation_counter(void) {
    return atomic_load(&g_allocation_counter);
}
#else
void nextimage_increment_alloc_counter(void) {}
void nextimage_decrement_alloc_counter(void) {}
#endif

// 内部用: エラーメッセージを設定
void nextimage_set_error(const char* format, ...) {
    va_list args;
    va_start(args, format);
    vsnprintf(g_error_buffer, sizeof(g_error_buffer), format, args);
    va_end(args);
}

// エラーメッセージ取得
const char* nextimage_last_error_message(void) {
    if (g_error_buffer[0] == '\0') {
        return NULL;
    }
    return g_error_buffer;
}

// エラーメッセージのクリア
void nextimage_clear_error(void) {
    g_error_buffer[0] = '\0';
}

// バッファの解放
void nextimage_free_buffer(NextImageBuffer* buffer) {
    if (buffer && buffer->data) {
        free(buffer->data);
        nextimage_decrement_alloc_counter();
        buffer->data = NULL;
        buffer->size = 0;
    }
}

// デコードバッファの解放
void nextimage_free_decode_buffer(NextImageDecodeBuffer* buffer) {
    if (!buffer || !buffer->owns_data) {
        return;
    }

    if (buffer->data) {
        free(buffer->data);
        nextimage_decrement_alloc_counter();
        buffer->data = NULL;
    }

    if (buffer->u_plane) {
        free(buffer->u_plane);
        nextimage_decrement_alloc_counter();
        buffer->u_plane = NULL;
    }

    if (buffer->v_plane) {
        free(buffer->v_plane);
        nextimage_decrement_alloc_counter();
        buffer->v_plane = NULL;
    }

    buffer->data_capacity = 0;
    buffer->data_size = 0;
    buffer->u_capacity = 0;
    buffer->u_size = 0;
    buffer->v_capacity = 0;
    buffer->v_size = 0;
    buffer->owns_data = 0;
}

// バージョン取得
const char* nextimage_version(void) {
    static char version[64];
    snprintf(version, sizeof(version), "%d.%d.%d",
             NEXTIMAGE_VERSION_MAJOR,
             NEXTIMAGE_VERSION_MINOR,
             NEXTIMAGE_VERSION_PATCH);
    return version;
}

// 内部ヘルパー: メモリ割り当て（カウンター付き）
void* nextimage_malloc(size_t size) {
    void* ptr = malloc(size);
    if (ptr) {
        nextimage_increment_alloc_counter();
    }
    return ptr;
}

void* nextimage_calloc(size_t nmemb, size_t size) {
    void* ptr = calloc(nmemb, size);
    if (ptr) {
        nextimage_increment_alloc_counter();
    }
    return ptr;
}

void* nextimage_realloc(void* ptr, size_t size) {
    int was_allocated = (ptr != NULL);
    void* new_ptr = realloc(ptr, size);

    if (new_ptr && !was_allocated) {
        // 新規割り当て
        nextimage_increment_alloc_counter();
    } else if (!new_ptr && was_allocated) {
        // 解放された
        nextimage_decrement_alloc_counter();
    }

    return new_ptr;
}

void nextimage_free(void* ptr) {
    if (ptr) {
        free(ptr);
        nextimage_decrement_alloc_counter();
    }
}
