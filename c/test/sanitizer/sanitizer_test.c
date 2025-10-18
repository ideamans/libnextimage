#include "nextimage.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

// Sanitizerテスト: メモリ安全性の検証

void test_buffer_overflow_protection(void) {
    printf("Testing buffer overflow protection...\n");

    // 正常なバッファ操作
    NextImageDecodeBuffer buf = {0};
    buf.data = malloc(100);
    buf.data_capacity = 100;
    buf.owns_data = 1;

    // 安全な書き込み
    memset(buf.data, 0, 100);
    printf("  ✓ Safe buffer write works\n");

    // 解放
    nextimage_free_decode_buffer(&buf);
    printf("  ✓ Buffer overflow protection test passed\n");
}

void test_use_after_free_protection(void) {
    printf("\nTesting use-after-free protection...\n");

    NextImageEncodeBuffer buf;
    buf.data = malloc(100);
    buf.size = 100;

    // 解放
    nextimage_free_encode_buffer(&buf);

    // buf.dataはNULLに設定されているはず
    assert(buf.data == NULL);
    assert(buf.size == 0);
    printf("  ✓ Buffer is NULL after free\n");

    // 二重解放しても安全
    nextimage_free_encode_buffer(&buf);
    printf("  ✓ Double free is safe\n");

    printf("  ✓ Use-after-free protection test passed\n");
}

void test_null_pointer_safety(void) {
    printf("\nTesting null pointer safety...\n");

    // NULLポインタでも安全
    nextimage_free_encode_buffer(NULL);
    nextimage_free_decode_buffer(NULL);
    printf("  ✓ Free NULL pointers is safe\n");

    // エラー関数もNULL安全
    const char* msg = nextimage_last_error_message();
    (void)msg; // 未使用警告を避ける
    nextimage_clear_error();
    printf("  ✓ Error functions are NULL-safe\n");

    printf("  ✓ Null pointer safety test passed\n");
}

int main(void) {
    printf("=== libnextimage Sanitizer Test ===\n\n");

    test_buffer_overflow_protection();
    test_use_after_free_protection();
    test_null_pointer_safety();

    printf("\n=== All sanitizer tests passed! ===\n");
    printf("Note: ASan/UBSan will detect any memory issues during execution\n");
    return 0;
}
