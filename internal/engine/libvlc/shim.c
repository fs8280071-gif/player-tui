#include "shim.h"

#include <pthread.h>
#include <stdlib.h>
#include <string.h>

struct pt_vlc_ctx {
    void *buf;
    size_t size;
    unsigned w;
    unsigned h;
    unsigned pitches[3];
    uintptr_t handle;
    pthread_mutex_t mu;
    uint64_t seq;
};

static unsigned pt_setup(void **opaque, char *chroma, unsigned *width,
                         unsigned *height, unsigned *pitches, unsigned *lines)
{
    pt_vlc_ctx *ctx = (pt_vlc_ctx *)*opaque;
    if (ctx == NULL) {
        return 0;
    }
    unsigned w = *width;
    unsigned h = *height;
    if (w == 0 || h == 0) {
        return 0;
    }
    unsigned pitch_y = w;
    unsigned pitch_uv = (w + 1) / 2;
    unsigned lines_uv = (h + 1) / 2;
    size_t size = (size_t)pitch_y * h + 2 * (size_t)pitch_uv * lines_uv;

    void *nbuf = malloc(size);
    if (nbuf == NULL) {
        return 0;
    }

    memcpy(chroma, "I420", 4);

    pthread_mutex_lock(&ctx->mu);
    void *old = ctx->buf;
    ctx->buf = nbuf;
    ctx->size = size;
    ctx->w = w;
    ctx->h = h;
    ctx->pitches[0] = pitch_y;
    ctx->pitches[1] = pitch_uv;
    ctx->pitches[2] = pitch_uv;
    pthread_mutex_unlock(&ctx->mu);
    free(old);

    pitches[0] = pitch_y;
    pitches[1] = pitch_uv;
    pitches[2] = pitch_uv;
    lines[0] = h;
    lines[1] = lines_uv;
    lines[2] = lines_uv;
    return 3;
}

static void pt_cleanup(void *opaque)
{
    pt_vlc_ctx *ctx = (pt_vlc_ctx *)opaque;
    if (ctx == NULL) {
        return;
    }
    pthread_mutex_lock(&ctx->mu);
    void *old = ctx->buf;
    ctx->buf = NULL;
    ctx->size = 0;
    ctx->w = 0;
    ctx->h = 0;
    pthread_mutex_unlock(&ctx->mu);
    free(old);
}

static void *pt_lock(void *opaque, void **planes)
{
    pt_vlc_ctx *ctx = (pt_vlc_ctx *)opaque;
    pthread_mutex_lock(&ctx->mu);
    if (ctx->buf == NULL) {
        planes[0] = NULL;
        planes[1] = NULL;
        planes[2] = NULL;
        return NULL;
    }
    unsigned lines_uv = (ctx->h + 1) / 2;
    uint8_t *y = (uint8_t *)ctx->buf;
    uint8_t *u = y + (size_t)ctx->pitches[0] * ctx->h;
    uint8_t *v = u + (size_t)ctx->pitches[1] * lines_uv;
    planes[0] = y;
    planes[1] = u;
    planes[2] = v;
    return ctx->buf;
}

static void pt_unlock(void *opaque, void *picture, void *const *planes)
{
    (void)picture;
    (void)planes;
    pt_vlc_ctx *ctx = (pt_vlc_ctx *)opaque;
    pthread_mutex_unlock(&ctx->mu);
}

static void pt_display(void *opaque, void *picture)
{
    (void)picture;
    pt_vlc_ctx *ctx = (pt_vlc_ctx *)opaque;
    __atomic_add_fetch(&ctx->seq, 1, __ATOMIC_RELEASE);
}

pt_vlc_ctx *pt_ctx_new(uintptr_t handle)
{
    pt_vlc_ctx *ctx = (pt_vlc_ctx *)calloc(1, sizeof(*ctx));
    if (ctx == NULL) {
        return NULL;
    }
    ctx->handle = handle;
    if (pthread_mutex_init(&ctx->mu, NULL) != 0) {
        free(ctx);
        return NULL;
    }
    return ctx;
}

void pt_ctx_free(pt_vlc_ctx *ctx)
{
    if (ctx == NULL) {
        return;
    }
    pthread_mutex_lock(&ctx->mu);
    free(ctx->buf);
    ctx->buf = NULL;
    pthread_mutex_unlock(&ctx->mu);
    pthread_mutex_destroy(&ctx->mu);
    free(ctx);
}

int pt_add_subtitle(libvlc_media_player_t *mp, const char *path)
{
    return libvlc_media_player_add_slave(mp, libvlc_media_slave_type_subtitle,
                                         path, true);
}

int pt_ctx_attach(pt_vlc_ctx *ctx, libvlc_media_player_t *mp)
{
    if (ctx == NULL || mp == NULL) {
        return -1;
    }
    libvlc_video_set_callbacks(mp, pt_lock, pt_unlock, pt_display, ctx);
    libvlc_video_set_format_callbacks(mp, pt_setup, pt_cleanup);
    return 0;
}

void pt_ctx_detach(pt_vlc_ctx *ctx, libvlc_media_player_t *mp)
{
    (void)ctx;
    if (mp == NULL) {
        return;
    }
    libvlc_video_set_callbacks(mp, NULL, NULL, NULL, NULL);
    libvlc_video_set_format_callbacks(mp, NULL, NULL);
}

void pt_ctx_lock(pt_vlc_ctx *ctx)
{
    pthread_mutex_lock(&ctx->mu);
}

void pt_ctx_unlock(pt_vlc_ctx *ctx)
{
    pthread_mutex_unlock(&ctx->mu);
}

uint64_t pt_ctx_seq(pt_vlc_ctx *ctx)
{
    return __atomic_load_n(&ctx->seq, __ATOMIC_ACQUIRE);
}

const uint8_t *pt_ctx_data(pt_vlc_ctx *ctx, unsigned *w, unsigned *h,
                           unsigned *pitch_y, unsigned *pitch_u,
                           unsigned *pitch_v)
{
    *w = ctx->w;
    *h = ctx->h;
    *pitch_y = ctx->pitches[0];
    *pitch_u = ctx->pitches[1];
    *pitch_v = ctx->pitches[2];
    return (const uint8_t *)ctx->buf;
}
