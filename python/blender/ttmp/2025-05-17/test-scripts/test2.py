import bpy
import os

def sequence_disco_fever():
    """Analyze the sequence editor and print information about strips with a disco twist."""
    scene = bpy.context.scene
    
    # Check if sequence editor exists
    if not scene.sequence_editor:
        print("🎭 No dance floor found! Creating one... 🎭")
        scene.sequence_editor_create()
        return "🎬 Dance floor installed, but no dancers found! 🎬"
    
    # Count strips
    strips = scene.sequence_editor.sequences
    total_strips = len(strips)
    
    if total_strips == 0:
        return "😱 Oh no! The dance floor is empty! Add some strips to party! 😱"
    
    # Categorize strips by type
    strip_types = {}
    for strip in strips:
        strip_type = strip.type
        if strip_type not in strip_types:
            strip_types[strip_type] = []
        strip_types[strip_type].append(strip.name)
    
    # Build report with funny names
    report = f"🎉 Total party people: {total_strips} strips 🎉\n\n"
    report += "💃 Party crew breakdown 🕺:\n"
    
    for strip_type, names in strip_types.items():
        if strip_type == "MOVIE":
            type_name = "Movie Stars"
        elif strip_type == "SOUND":
            type_name = "Beat Droppers"
        elif strip_type == "IMAGE":
            type_name = "Snapshot Divas"
        else:
            type_name = strip_type + " Performers"
            
        report += f"- {type_name}: {len(names)} dancers\n"
        for name in names:
            report += f"  - {name}\n"
    
    return report

def clip_teleporter_5000(filepath, channel=1, start_frame=1, name=None):
    """
    Teleport a media file into the sequence editor universe.
    
    Args:
        filepath (str): Secret coordinates to the media file
        channel (int, optional): Dimensional layer to place the media. Defaults to 1.
        start_frame (int, optional): Time coordinate to begin existence. Defaults to 1.
        name (str, optional): Secret identity for the strip. Defaults to None.
        
    Returns:
        The newly materialized strip entity or None if teleportation failed
    """
    if not os.path.exists(filepath):
        print(f"🚨 ERROR: File {filepath} not found in this dimension! 🚨")
        return None
        
    scene = bpy.context.scene
    
    # Ensure sequence editor exists
    if not scene.sequence_editor:
        print("🌌 Creating a new dimension for your media... 🌌")
        scene.sequence_editor_create()
    
    # Get file extension
    _, ext = os.path.splitext(filepath)
    ext = ext.lower()
    
    # Add the appropriate strip type with funny messages
    if ext in ['.mp4', '.avi', '.mov', '.mkv', '.flv', '.webm']:
        strip = scene.sequence_editor.sequences.new_movie(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"🎬 Movie star has entered the stage: {strip.name} 🎬")
    elif ext in ['.mp3', '.wav', '.ogg', '.flac']:
        strip = scene.sequence_editor.sequences.new_sound(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"🎵 Sound wizard has cast their spell: {strip.name} 🎵")
    elif ext in ['.png', '.jpg', '.jpeg', '.tiff', '.bmp']:
        strip = scene.sequence_editor.sequences.new_image(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"📸 Image ninja has appeared: {strip.name} 📸")
    else:
        print(f"❓ What sorcery is this? Unknown file type: {ext} ❓")
        return None
        
    return strip

def strip_zapper_deluxe(strip_name):
    """Delete a strip with dramatic flair."""
    scene = bpy.context.scene
    
    if not scene.sequence_editor:
        return "🤷‍♂️ No sequence editor found. Nothing to zap! 🤷‍♂️"
    
    strips = scene.sequence_editor.sequences
    for strip in strips:
        if strip.name == strip_name:
            strips.remove(strip)
            return f"💥 KAPOW! Strip '{strip_name}' has been vaporized! 💥"
    
    return f"🧐 Hmm, couldn't find '{strip_name}' anywhere. Did it escape? 🧐"

def mash_o_matic(strip1_name, strip2_name, transition_type="CROSS", duration=10):
    """Create a transition between two strips with pizzazz."""
    scene = bpy.context.scene
    
    if not scene.sequence_editor:
        return "🏜️ Wasteland detected! No sequence editor to mash in! 🏜️"
    
    strips = scene.sequence_editor.sequences
    strip1 = None
    strip2 = None
    
    for strip in strips:
        if strip.name == strip1_name:
            strip1 = strip
        elif strip.name == strip2_name:
            strip2 = strip
    
    if not strip1 or not strip2:
        return "🤔 Can't find one or both of your strips. Check your spelling! 🤔"
    
    # Create a transition
    try:
        if transition_type == "CROSS":
            effect = strips.new_effect(
                name=f"{strip1_name}_{strip2_name}_mashup",
                type="CROSS",
                channel=max(strip1.channel, strip2.channel) + 1,
                frame_start=max(strip1.frame_final_end - duration, strip2.frame_start),
                frame_end=max(strip1.frame_final_end, strip2.frame_start + duration)
            )
            effect.seq1 = strip1
            effect.seq2 = strip2
            return f"✨ Voilà! Created a magical transition between {strip1_name} and {strip2_name}! ✨"
    except Exception as e:
        return f"💔 Transition creation failed with error: {str(e)} 💔"

# Function to call if you just want to have fun
def party_time():
    """Start the party with a sequence analysis!"""
    print("🎊🎊🎊 LET'S GET THIS PARTY STARTED! 🎊🎊🎊")
    result = sequence_disco_fever()
    print(result)
    print("🎊🎊🎊 PARTY ON, DUDES! 🎊🎊🎊")
    return result

# Only run this when file is executed directly
if __name__ == "__main__":
    party_time() 