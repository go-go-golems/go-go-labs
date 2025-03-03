## Film Processing Documentation

Added comprehensive documentation for RECORD Speed Fixer based on product label information. The guide includes detailed mixing instructions for different applications (film fixing, print fixing, fast print fixing, and hardening print fixing), usage guidelines, capacity information, shelf life details, and safety precautions.

## Stop Bath Documentation

Added comprehensive documentation for BLOCK Stop Bath based on product label information. The guide includes information about advantages over conventional stop baths, preparation instructions, usage guidelines for films and prints, capacity information with color indicator details, instructions for creating a hardening stop bath, shelf life, and safety precautions. 

## Bootstrap Integration in Film Development Timer

Enhanced the Film Development Timer application by integrating Bootstrap CSS directly into the shadow DOM of all Lit components. Created a base `BootstrapLitElement` class that automatically injects the Bootstrap stylesheet into each component's shadow root, ensuring proper styling while maintaining component encapsulation. This approach preserves the benefits of shadow DOM isolation while allowing components to use Bootstrap's styling classes. 

## Film Development Timer Usability Improvements

Enhanced the Film Development Timer application with several usability improvements:
- Made all optional steps unchecked by default to allow users to explicitly choose which steps they want to include
- Added a "Restart Step" button to allow users to restart the current processing step if needed
- Added a "+30s" button to extend the current step's duration by 30 seconds for more flexibility during development 