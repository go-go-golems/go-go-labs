import bpy # type: ignore
import os
from mathutils import * # type: ignore

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def print_scene_fps_info(scene=None):
    """Print detailed FPS and timing information for a scene."""
    if scene is None:
        scene = bpy.context.scene
    
    # --- FPS / duration ------------------------------------------------------
    fps = scene.render.fps / scene.render.fps_base
    print(f"\nFPS Information for Scene: {scene.name}")
    print("=" * 50)
    print(f"Frame Rate Base: {scene.render.fps}/{scene.render.fps_base}")
    print(f"Effective FPS : {fps:.2f}")
    
    # --- Time-remapping ------------------------------------------------------
    rm_old = scene.render.frame_map_old
    rm_new = scene.render.frame_map_new
    remap_enabled = (rm_old != 1 or rm_new != 1)
    print(f"\nTime Remapping : {'Enabled' if remap_enabled else 'Disabled'} "
          f"(old={rm_old}, new={rm_new})")
    
    # --- Frame range / duration ---------------------------------------------
    print(f"\nFrame Range    : {scene.frame_start}-{scene.frame_end}")
    total = scene.frame_end - scene.frame_start + 1
    dur_sec = total / fps
    print(f"Duration       : {int(dur_sec//60)}m {dur_sec%60:.2f}s "
          f"({total} frames)")
    
    # --- VSE-specific info ---------------------------------------------------
    se = scene.sequence_editor
    if se:
        print("\nVSE Info       :")
        print(f"Proxy Storage  : {se.proxy_storage!s:<9}  dir={se.proxy_dir or '(default)'}")
        # Strip-level proxy flags
        for s in se.sequences_all:
            if getattr(s, 'use_proxy', False):
                p = s.proxy
                built = [f"{pcent}%" for pcent, flag in
                         ((25, p.build_25), (50, p.build_50),
                          (75, p.build_75), (100, p.build_100))
                         if flag]
                built = ', '.join(built) or '—'
                print(f"  • {s.name:<24} uses proxy ({built})")
        # Print remapping info for each strip
        print("\nStrip Remapping Info:")
        for s in se.sequences_all:
            remapped = False
            # Check for speed effect strip
            if s.type == 'SPEED':
                remapped = True
            # Check for time remapping modifiers (e.g., Speed Control effect)
            if hasattr(s, 'modifiers'):
                for mod in s.modifiers:
                    if mod.type == 'SPEED':
                        remapped = True
            # Check for speed_factor property (for some strip types)
            if hasattr(s, 'speed_factor') and getattr(s, 'speed_factor', 1.0) != 1.0:
                remapped = True

            # Scene settings (moved up for diag)
            scene_fps = scene.render.fps / scene.render.fps_base
            scene_audio_sr = getattr(scene.render, 'ffmpeg', None)
            if scene_audio_sr:
                scene_audio_sr = getattr(scene.render.ffmpeg, 'audio_mixrate', None)
            else:
                scene_audio_sr = None

            # Diagnostics
            diag = clip_diagnostics(s, scene_fps)
            if diag:
                print(f"    [DEBUG] {s.name} diagnostics: {diag}")

            # Determine strip's effective FPS and SR from diagnostics
            effective_strip_video_fps = None
            effective_strip_audio_sr = None

            if diag:
                if diag.get('kind') == 'video':
                    effective_strip_video_fps = diag.get('src_fps')
                elif diag.get('kind') == 'audio':
                    effective_strip_audio_sr = diag.get('sample_rt')
            
            # Length in frames and seconds
            frames = getattr(s, 'frame_final_duration', None)
            seconds_to_display = None
            if frames is not None:
                # Prefer timeline_s from diag if available, as it directly reflects timeline duration
                if diag and diag.get('timeline_s') is not None:
                    seconds_to_display = diag.get('timeline_s')
                else: # Fallback if diag or timeline_s not available
                    seconds_to_display = frames / scene_fps

            info = f"  • {s.name:<24} remapped: {'YES' if remapped else 'NO'}"
            if effective_strip_audio_sr:
                info += f" | audio sample rate: {effective_strip_audio_sr} Hz"
                if scene_audio_sr:
                    info += f" (scene: {scene_audio_sr} Hz)"
            if effective_strip_video_fps:
                info += f" | video fps: {effective_strip_video_fps:.2f} (scene: {scene_fps:.2f})"
            
            if frames is not None:
                info += f" | length: {frames} frames"
                if seconds_to_display is not None:
                    info += f" ({seconds_to_display:.2f} s)"
            
            # Warnings
            if effective_strip_video_fps and abs(effective_strip_video_fps - scene_fps) > 0.1:
                info += f"  ⚠️  FPS mismatch!"
            if effective_strip_audio_sr and scene_audio_sr and effective_strip_audio_sr != scene_audio_sr:
                info += f"  ⚠️  SR mismatch!"
            print(info)

def clip_diagnostics(strip, scene_fps):
    # Only MOVIE and SOUND strips are interesting here
    if strip.type not in {'MOVIE', 'SOUND'}:
        return None

    if strip.type == 'MOVIE':
        src_fps = None
        if hasattr(strip.elements[0], 'orig_fps') and strip.elements[0].orig_fps:
            src_fps = strip.elements[0].orig_fps
        elif hasattr(strip, 'fps'):
            src_fps = strip.fps
        src_secs = strip.frame_final_duration / scene_fps      # timeline seconds
        true_secs = strip.frame_final_duration / src_fps if src_fps else None  # what the file "expects"
        return {
            'kind'      : 'video',
            'src_fps'   : src_fps,
            'timeline_s': src_secs,
            'true_s'    : true_secs,
            'delta_s'   : src_secs - true_secs if true_secs is not None else None
        }

    if strip.type == 'SOUND':
        sr   = getattr(strip.sound, 'sample_rate', None)
        samp = getattr(strip.sound, 'frame_duration', None)    # total audio samples
        true_secs = samp / sr if (samp and sr) else None
        tl_secs   = strip.frame_final_duration / scene_fps
        return {
            'kind'      : 'audio',
            'sample_rt' : sr,
            'timeline_s': tl_secs,
            'true_s'    : true_secs,
            'delta_s'   : tl_secs - true_secs if true_secs is not None else None
        }

def main():
    """Main function to test scene speeds and timing."""
    print("Blender Scene Speed Test")
    print("======================")
    
    # Test active scene
    print_scene_fps_info()
    
    # Test all scenes
    for scene in D.scenes:
        if scene != C.scene:
            print_scene_fps_info(scene)

main()