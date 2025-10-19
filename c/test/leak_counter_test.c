#include "nextimage.h"
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#ifdef NEXTIMAGE_DEBUG

void test_leak_counter(void) {
    printf("Testing leak counter...\n");

    int64_t initial = nextimage_allocation_counter();
    printf("  Initial counter: %lld\n", (long long)initial);

    // エンコードバッファを手動割り当て
    NextImageBuffer enc_buf;
    enc_buf.data = malloc(1024);
    enc_buf.size = 1024;

    // カウンターは変わらない（malloc直接呼び出しのため）
    int64_t after_malloc = nextimage_allocation_counter();
    assert(after_malloc == initial);
    printf("  ✓ Direct malloc doesn't affect counter\n");

    // 解放
    free(enc_buf.data);
    enc_buf.data = NULL;

    // デコードバッファを手動割り当て
    NextImageDecodeBuffer dec_buf = {0};
    dec_buf.data = malloc(1024);
    dec_buf.u_plane = malloc(256);
    dec_buf.v_plane = malloc(256);
    dec_buf.owns_data = 1;

    // カウンターは変わらない
    after_malloc = nextimage_allocation_counter();
    assert(after_malloc == initial);
    printf("  ✓ Direct malloc for decode buffer doesn't affect counter\n");

    // 解放
    free(dec_buf.data);
    free(dec_buf.u_plane);
    free(dec_buf.v_plane);

    int64_t final = nextimage_allocation_counter();
    assert(final == initial);
    printf("  ✓ Final counter matches initial: %lld\n", (long long)final);

    printf("  ✓ Leak counter test passed\n");
    printf("  Note: Internal malloc/free tracking will be tested with actual encode/decode\n");
}

int main(void) {
    printf("=== libnextimage Leak Counter Test ===\n\n");

    test_leak_counter();

    printf("\n=== All leak counter tests passed! ===\n");
    return 0;
}

#else

int main(void) {
    printf("=== libnextimage Leak Counter Test ===\n\n");
    printf("NEXTIMAGE_DEBUG not defined - leak counter is not available\n");
    printf("This test requires a debug build\n");
    return 0;
}

#endif
