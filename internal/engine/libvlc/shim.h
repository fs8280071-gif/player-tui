#ifndef PLAYER_TUI_VLC_SHIM_H
#define PLAYER_TUI_VLC_SHIM_H

#include <vlc/vlc.h>
#include <stddef.h>
#include <stdint.h>

typedef struct pt_vlc_ctx pt_vlc_ctx;

pt_vlc_ctx *pt_ctx_new(uintptr_t handle);
void pt_ctx_free(pt_vlc_ctx *ctx);
int pt_ctx_attach(pt_vlc_ctx *ctx, libvlc_media_player_t *mp);
int pt_add_subtitle(libvlc_media_player_t *mp, const char *path);
void pt_ctx_detach(pt_vlc_ctx *ctx, libvlc_media_player_t *mp);

/* pt_ctx_lock/pt_ctx_unlock protect the C-owned frame buffer. */
void pt_ctx_lock(pt_vlc_ctx *ctx);
void pt_ctx_unlock(pt_vlc_ctx *ctx);
uint64_t pt_ctx_seq(pt_vlc_ctx *ctx);
const uint8_t *pt_ctx_data(pt_vlc_ctx *ctx, unsigned *w, unsigned *h,
                           unsigned *pitch_y, unsigned *pitch_u,
                           unsigned *pitch_v);

#endif
