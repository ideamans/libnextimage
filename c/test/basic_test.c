#include "nextimage.h"
#include <stdio.h>
#include <string.h>
#include <assert.h>

void test_version(void) {
    printf("Testing version...\n");
    const char* version = nextimage_version();
    assert(version != NULL);
    assert(strlen(version) > 0);
    printf("  Version: %s\n", version);
    printf("  ✓ Version test passed\n");
}

void test_error_handling(void) {
    printf("\nTesting error handling...\n");

    // 初期状態ではエラーメッセージはNULL
    const char* msg = nextimage_last_error_message();
    assert(msg == NULL);
    printf("  ✓ Initial error message is NULL\n");

    // エラーをクリア
    nextimage_clear_error();
    msg = nextimage_last_error_message();
    assert(msg == NULL);
    printf("  ✓ Clear error works\n");

    printf("  ✓ Error handling test passed\n");
}

void test_buffer_allocation(void) {
    printf("\nTesting buffer allocation...\n");

    // エンコードバッファのテスト
    NextImageEncodeBuffer enc_buf = {0};
    enc_buf.data = NULL;
    enc_buf.size = 0;

    // 解放（NULLでも安全）
    nextimage_free_encode_buffer(&enc_buf);
    printf("  ✓ Free NULL encode buffer is safe\n");

    // デコードバッファのテスト
    NextImageDecodeBuffer dec_buf = {0};
    dec_buf.data = NULL;
    dec_buf.u_plane = NULL;
    dec_buf.v_plane = NULL;
    dec_buf.owns_data = 0;

    // 解放（owns_data = 0なので何もしない）
    nextimage_free_decode_buffer(&dec_buf);
    printf("  ✓ Free non-owned decode buffer is safe\n");

    printf("  ✓ Buffer allocation test passed\n");
}

int main(void) {
    printf("=== libnextimage Basic Test ===\n\n");

    test_version();
    test_error_handling();
    test_buffer_allocation();

    printf("\n=== All tests passed! ===\n");
    return 0;
}
