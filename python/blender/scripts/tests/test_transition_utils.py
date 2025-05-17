# Test script for transition_utils.py

import bpy
import sys
import os
import importlib

# Add script directories to path
scripts_dir = os.path.dirname(os.path.abspath(__file__))
project_dir = os.path.dirname(scripts_dir)
utils_dir = os.path.join(project_dir, 'utils')

for path in [scripts_dir, project_dir, utils_dir]:
    if path not in sys.path:
        sys.path.append(path)

# Import the VSE utilities and transition utilities
from utils import vse_utils
importlib.reload(vse_utils)
from utils import transition_utils
importlib.reload(transition_utils)

# Common shortcuts
D = bpy.data
C = bpy.context

def setup_test_scene():
    """Set up a test scene with two video clips for transition testing."""
    print("\n---- Setting up test scene ----")
    
    scene = vse_utils.get_active_scene()
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Clear any existing strips
    vse_utils.clear_all_strips(seq_editor)
    
    # Find test media directory
    test_media_dir = vse_utils.find_test_media_dir()
    
    # Add two test clips with a small gap between them
    video_files = [
        "SampleVideo_1280x720_1mb.mp4",
        "SampleVideo_1280x720_2mb.mp4"
    ]
    
    clips = []
    frame_start = 1
    
    for i, filename in enumerate(video_files):
        filepath = os.path.join(test_media_dir, filename)
        if os.path.exists(filepath):
            # Add video on channel 1
            video = vse_utils.add_movie_strip(
                seq_editor,
                filepath,
                channel=1,
                frame_start=frame_start,
                name=f"Video{i+1}"
            )
            
            # Add audio on channel 2
            audio = vse_utils.add_sound_strip(
                seq_editor,
                filepath,
                channel=2,
                frame_start=frame_start,
                name=f"Audio{i+1}"
            )
            
            clips.append((video, audio))
            
            # Update frame_start for next clip (with a small gap)
            frame_start = video.frame_final_end + 30
        else:
            print(f"Warning: Test file not found: {filepath}")
    
    vse_utils.print_sequence_info(seq_editor, "Initial Test Scene")
    return scene, seq_editor, clips

def test_create_crossfade():
    """Test the create_crossfade function."""
    print("\n==== Testing create_crossfade ====\n")
    
    scene, seq_editor, clips = setup_test_scene()
    
    if len(clips) < 2:
        print("Error: Need at least 2 clips for transition testing")
        return False
    
    # Get the video strips
    video1, audio1 = clips[0]
    video2, audio2 = clips[1]
    
    # Move the second clip to overlap with the first
    overlap_frames = 24  # 1 second at 24fps
    video2.frame_start = video1.frame_final_end - overlap_frames
    audio2.frame_start = video2.frame_start
    
    print("\nAfter adjusting clip positions for overlap:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Create crossfade
    transition = transition_utils.create_crossfade(
        seq_editor,
        video1,
        video2,
        overlap_frames
    )
    
    print("\nAfter adding crossfade:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Check if the transition was created successfully
    if transition and transition.type == 'CROSS':
        print("Crossfade created successfully!")
        return True
    else:
        print("Failed to create crossfade")
        return False

def test_create_gamma_crossfade():
    """Test the create_gamma_crossfade function."""
    print("\n==== Testing create_gamma_crossfade ====\n")
    
    scene, seq_editor, clips = setup_test_scene()
    
    if len(clips) < 2:
        print("Error: Need at least 2 clips for transition testing")
        return False
    
    # Get the video strips
    video1, audio1 = clips[0]
    video2, audio2 = clips[1]
    
    # Move the second clip to overlap with the first
    overlap_frames = 30  # 1.25 seconds at 24fps
    video2.frame_start = video1.frame_final_end - overlap_frames
    audio2.frame_start = video2.frame_start
    
    print("\nAfter adjusting clip positions for overlap:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Create gamma crossfade
    transition = transition_utils.create_gamma_crossfade(
        seq_editor,
        video1,
        video2,
        overlap_frames,
        channel=3  # Test explicit channel specification
    )
    
    print("\nAfter adding gamma crossfade:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Check if the transition was created successfully
    if transition and transition.type == 'GAMMA_CROSS' and transition.channel == 3:
        print("Gamma crossfade created successfully on channel 3!")
        return True
    else:
        print("Failed to create gamma crossfade")
        return False

def test_create_wipe():
    """Test the create_wipe function."""
    print("\n==== Testing create_wipe ====\n")
    
    scene, seq_editor, clips = setup_test_scene()
    
    if len(clips) < 2:
        print("Error: Need at least 2 clips for transition testing")
        return False
    
    # Get the video strips
    video1, audio1 = clips[0]
    video2, audio2 = clips[1]
    
    # Move the second clip to overlap with the first
    overlap_frames = 36  # 1.5 seconds at 24fps
    video2.frame_start = video1.frame_final_end - overlap_frames
    audio2.frame_start = video2.frame_start
    
    print("\nAfter adjusting clip positions for overlap:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Create wipe transition with custom parameters
    transition = transition_utils.create_wipe(
        seq_editor,
        video1,
        video2,
        overlap_frames,
        wipe_type='CLOCK',  # Test clock wipe
        angle=0.7,  # Non-zero angle
        channel=4    # Test explicit channel
    )
    
    print("\nAfter adding wipe transition:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Check if the transition was created successfully with correct parameters
    if (transition and transition.type == 'WIPE' and 
            transition.channel == 4 and 
            transition.transition_type == 'CLOCK' and
            abs(transition.angle - 0.7) < 0.01):
        print("Wipe transition created successfully with correct parameters!")
        return True
    else:
        print("Failed to create wipe transition or parameters incorrect")
        return False

def test_create_audio_fade():
    """Test the create_audio_fade function for both fade in and fade out."""
    print("\n==== Testing create_audio_fade ====\n")
    
    scene, seq_editor, clips = setup_test_scene()
    
    if not clips:
        print("Error: No clips available for audio fade testing")
        return False
    
    # Get the audio strip from the first clip
    video1, audio1 = clips[0]
    
    # Test fade in
    print("\nTesting audio fade IN:")
    success_in = transition_utils.create_audio_fade(
        audio1, 
        fade_type='IN', 
        duration_frames=20
    )
    
    # Check for keyframes on volume
    has_keyframes = False
    scene = bpy.context.scene
    if scene.animation_data and scene.animation_data.action:
        for fc in scene.animation_data.action.fcurves:
            if fc.data_path.startswith('sequence_editor.sequences_all[') and fc.data_path.endswith('].volume'):
                if audio1.name in fc.data_path:
                    has_keyframes = True
                    break
    
    print(f"Audio fade IN {'succeeded' if success_in else 'failed'}")
    print(f"Has volume keyframes: {has_keyframes}")
    
    # Set up a new scene for testing fade out
    scene, seq_editor, clips = setup_test_scene()
    video1, audio1 = clips[0]
    
    # Test fade out
    print("\nTesting audio fade OUT:")
    success_out = transition_utils.create_audio_fade(
        audio1, 
        fade_type='OUT', 
        duration_frames=24
    )
    
    # Check for keyframes again
    has_keyframes = False
    scene = bpy.context.scene
    if scene.animation_data and scene.animation_data.action:
        for fc in scene.animation_data.action.fcurves:
            if fc.data_path.startswith('sequence_editor.sequences_all[') and fc.data_path.endswith('].volume'):
                if audio1.name in fc.data_path:
                    has_keyframes = True
                    break
    
    print(f"Audio fade OUT {'succeeded' if success_out else 'failed'}")
    print(f"Has volume keyframes: {has_keyframes}")
    
    return success_in and success_out

def test_create_audio_crossfade():
    """Test the create_audio_crossfade function."""
    print("\n==== Testing create_audio_crossfade ====\n")
    
    scene, seq_editor, clips = setup_test_scene()
    
    if len(clips) < 2:
        print("Error: Need at least 2 clips for audio crossfade testing")
        return False
    
    # Get the audio strips
    video1, audio1 = clips[0]
    video2, audio2 = clips[1]
    
    # No need to move clips yet - we'll let the function handle it
    
    # Create audio crossfade
    overlap_frames = 24
    success = transition_utils.create_audio_crossfade(
        audio1,
        audio2,
        overlap_frames
    )
    
    print("\nAfter adding audio crossfade:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Check for keyframes on both audio strips
    has_keyframes1 = False
    has_keyframes2 = False
    
    # Check the scene's animation data since strip keyframes are stored there
    scene = bpy.context.scene
    if scene.animation_data and scene.animation_data.action:
        for fc in scene.animation_data.action.fcurves:
            if fc.data_path.startswith('sequence_editor.sequences_all[') and fc.data_path.endswith('].volume'):
                # Check which strip this is for
                if audio1.name in fc.data_path:
                    has_keyframes1 = True
                elif audio2.name in fc.data_path:
                    has_keyframes2 = True
    
    print(f"Audio1 has volume keyframes: {has_keyframes1}")
    print(f"Audio2 has volume keyframes: {has_keyframes2}")
    
    return success and has_keyframes1 and has_keyframes2

def test_create_fade_to_color():
    """Test the create_fade_to_color function for both fade in and fade out."""
    print("\n==== Testing create_fade_to_color ====\n")
    
    # Test fade in from black
    print("\nTesting fade in from black:")
    scene, seq_editor, clips = setup_test_scene()
    
    if not clips:
        print("Error: No clips available for fade to color testing")
        return False
    
    # Get the video strip from the first clip
    video1, audio1 = clips[0]
    
    # Create fade in from black
    fade_in = transition_utils.create_fade_to_color(
        seq_editor,
        video1,
        fade_duration=20,
        fade_type='IN',
        color=(0, 0, 0)  # Black
    )
    
    print("\nAfter adding fade in from black:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Test fade out to white
    print("\nTesting fade out to white:")
    scene, seq_editor, clips = setup_test_scene()
    video1, audio1 = clips[0]
    
    # Create fade out to white
    fade_out = transition_utils.create_fade_to_color(
        seq_editor,
        video1,
        fade_duration=24,
        fade_type='OUT',
        color=(1, 1, 1)  # White
    )
    
    print("\nAfter adding fade out to white:")
    vse_utils.print_sequence_info(seq_editor)
    
    # Check if both transitions were created successfully
    return fade_in is not None and fade_out is not None

# Main function to run all tests
def run_all_tests():
    """Run all transition_utils tests."""
    print("\n===== RUNNING ALL TRANSITION UTILS TESTS =====\n")
    
    tests = [
        test_create_crossfade,
        test_create_gamma_crossfade,
        test_create_wipe,
        test_create_audio_fade,
        test_create_audio_crossfade,
        test_create_fade_to_color
    ]
    
    results = {}
    
    for test_func in tests:
        test_name = test_func.__name__
        try:
            success = test_func()
            results[test_name] = success
        except Exception as e:
            print(f"Exception in {test_name}: {e}")
            results[test_name] = False
    
    # Print summary
    print("\n===== TEST RESULTS SUMMARY =====\n")
    for test_name, success in results.items():
        status = "PASSED" if success else "FAILED"
        print(f"{test_name}: {status}")

# Individual test runners
def run_test_crossfade():
    test_create_crossfade()

def run_test_gamma_crossfade():
    test_create_gamma_crossfade()

def run_test_wipe():
    test_create_wipe()

def run_test_audio_fade():
    test_create_audio_fade()

def run_test_audio_crossfade():
    test_create_audio_crossfade()

def run_test_fade_to_color():
    test_create_fade_to_color()


# Add a helper function to import and reload the test module from an external script
def run_file(test_name):
    """Run a specific test by name"""
    # Map of test names to functions
    test_map = {
        'crossfade': run_test_crossfade,
        'gamma_crossfade': run_test_gamma_crossfade,
        'wipe': run_test_wipe,
        'audio_fade': run_test_audio_fade,
        'audio_crossfade': run_test_audio_crossfade,
        'fade_to_color': run_test_fade_to_color,
        'all': run_all_tests
    }
    
    if test_name in test_map:
        test_map[test_name]()
    else:
        print(f"Unknown test: {test_name}")
        print(f"Available tests: {', '.join(test_map.keys())}")

# Run a specific test when this script is executed directly
if __name__ == "__main__":
    # Uncomment the test you want to run
    # run_test_crossfade()
    # run_test_gamma_crossfade()
    # run_test_wipe()
    # run_test_audio_fade()
    # run_test_audio_crossfade()
    # run_test_fade_to_color()
    run_all_tests()