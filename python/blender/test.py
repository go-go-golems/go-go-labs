import bpy
import os

def analyze_sequence_editor():
    """Analyze the sequence editor and print information about strips."""
    scene = bpy.context.scene
    
    # Check if sequence editor exists
    if not scene.sequence_editor:
        print("No sequence editor found. Creating one.")
        scene.sequence_editor_create()
        return "Sequence editor created, but no strips found."
    
    # Count strips
    strips = scene.sequence_editor.sequences
    total_strips = len(strips)
    
    if total_strips == 0:
        return "No strips found in the sequence editor."
    
    # Categorize strips by type
    strip_types = {}
    for strip in strips:
        strip_type = strip.type
        if strip_type not in strip_types:
            strip_types[strip_type] = []
        strip_types[strip_type].append(strip.name)
    
    # Build report
    report = f"Total strips: {total_strips}\n\n"
    report += "Strips by type:\n"
    
    for strip_type, names in strip_types.items():
        report += f"- {strip_type}: {len(names)} strips\n"
        for name in names:
            report += f"  - {name}\n"
    
    return report

def import_file(filepath, channel=1, start_frame=1, name=None):
    """
    Import a file into the sequence editor.
    
    Args:
        filepath (str): The absolute path to the file
        channel (int, optional): The channel to place the strip on. Defaults to 1.
        start_frame (int, optional): The frame to start the strip at. Defaults to 1.
        name (str, optional): Custom name for the strip. Defaults to None (uses filename).
        
    Returns:
        The newly created strip object or None if failed
    """
    if not os.path.exists(filepath):
        print(f"Error: File {filepath} not found")
        return None
        
    scene = bpy.context.scene
    
    # Ensure sequence editor exists
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    # Get file extension
    _, ext = os.path.splitext(filepath)
    ext = ext.lower()
    
    # Add the appropriate strip type
    if ext in ['.mp4', '.avi', '.mov', '.mkv', '.flv', '.webm']:
        strip = scene.sequence_editor.sequences.new_movie(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"Added movie strip: {strip.name}")
    elif ext in ['.mp3', '.wav', '.ogg', '.flac']:
        strip = scene.sequence_editor.sequences.new_sound(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"Added sound strip: {strip.name}")
    elif ext in ['.png', '.jpg', '.jpeg', '.tiff', '.bmp']:
        strip = scene.sequence_editor.sequences.new_image(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"Added image strip: {strip.name}")
    else:
        print(f"Unsupported file type: {ext}")
        return None
        
    return strip

def dancing_strips_disco_party():
    """A function with a funny name that analyzes the sequence editor."""
    print("🎬 Let's get this party started with some sequence strips! 🎬")
    result = analyze_sequence_editor()
    print("🕺 Dance party analysis complete! 💃")
    return result

# Only run this when file is executed directly
if __name__ == "__main__":
    result = analyze_sequence_editor()
    print(result) 