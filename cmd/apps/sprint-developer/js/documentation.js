export const processDocumentation = {
  introduction: {
    title: "SPRINT Standard B&W Film Developer",
    description: "SPRINT Standard is the basic film developer in SPRINT's B&W processing system. It comes as a liquid concentrate that must be diluted before use. One liter of STANDARD concentrate will make 10 liters of working solution to develop at least 50 rolls of film, or enough replenished solution to develop 110 rolls."
  },
  preparation: {
    title: "Solution Preparation",
    steps: [
      {
        title: "Developer",
        ratio: "1:9",
        example: {
          concentrate: "100 ml STANDARD B&W Film Developer concentrate",
          water: "900 ml water",
          result: "1000 ml B&W Film Developer working solution"
        }
      },
      {
        title: "Stop Bath",
        ratio: "1:9",
        example: {
          concentrate: "100 ml BLOCK Stop Bath concentrate",
          water: "900 ml water",
          result: "1000 ml BLOCK Stop Bath working solution"
        }
      },
      {
        title: "Fixer",
        ratio: "2:8",
        example: {
          concentrate: "200 ml RECORD Speed Fixer concentrate",
          water: "800 ml water",
          hardener: "30 ml RECORD Alum Hardening Converter (optional)",
          result: "1030 ml Film Fixer working solution"
        }
      }
    ]
  },
  process: {
    title: "Development Process",
    steps: [
      {
        id: 1,
        name: "Water Pre-wet",
        duration: "1 minute",
        details: {
          purpose: "Prepares the film emulsion for even development",
          instructions: "Immerse film in water at development temperature",
          notes: "Use water at the same temperature as your developer solution"
        }
      },
      {
        id: 2,
        name: "Develop",
        duration: "See Development Chart",
        details: {
          purpose: "Converts exposed silver halides to metallic silver",
          instructions: "Use SPRINT Standard developer at chosen temperature",
          agitation: "Continuous for first minute, then 10-15 seconds every minute",
          notes: "Time depends on film type and temperature - refer to development chart"
        }
      },
      {
        id: 3,
        name: "Stop",
        duration: "1 minute",
        details: {
          purpose: "Halts development process immediately",
          instructions: "Use BLOCK Stop Bath solution",
          agitation: "Continuous for first minute",
          notes: "Prevents developer carryover and protects fixer",
          warnings: "Always use stop bath before fixing with alum hardener"
        }
      },
      {
        id: 4,
        name: "Fix",
        duration: "3 minutes",
        details: {
          purpose: "Removes unexposed silver halides, making image permanent",
          instructions: "Use RECORD Speed Fixer solution",
          agitation: "Continuous for first minute, then 10-15 seconds every minute",
          notes: "Can be used with optional hardener for increased emulsion durability"
        }
      },
      {
        id: 5,
        name: "Water Pre-wash",
        duration: "1 minute",
        details: {
          purpose: "Initial rinse to remove most of the fixer",
          instructions: "Use clean water at process temperature",
          notes: "Helps conserve fixer remover in next step"
        }
      },
      {
        id: 6,
        name: "Remove Fixer",
        duration: "3 minutes",
        details: {
          purpose: "Ensures complete removal of fixer for archival stability",
          instructions: "Use fixer remover solution",
          notes: "Critical step for long-term image permanence"
        }
      },
      {
        id: 7,
        name: "Water Wash",
        duration: "5 minutes",
        details: {
          purpose: "Final removal of all processing chemicals",
          instructions: "Use clean running water or multiple changes",
          notes: "Complete exchange of water three times per minute recommended"
        }
      },
      {
        id: 8,
        name: "Stabilize",
        duration: "1 minute",
        details: {
          purpose: "Prevents water spots and promotes even drying",
          instructions: "Use wetting agent solution",
          notes: "Do not skip this step - prevents drying marks"
        }
      },
      {
        id: 9,
        name: "Squeegee & Dry",
        duration: "-",
        details: {
          purpose: "Removes excess water and dries film",
          instructions: "Gently squeegee both sides, hang in dust-free area",
          notes: "Use film clips, avoid touching emulsion surface"
        }
      }
    ]
  },
  temperatureChart: {
    title: "Development Temperature Chart",
    description: "Find a convenient combination of Time and Temperature for development. Use any temperature 18-25°C / 64.5-77°F.",
    temperatures: [
      { celsius: 18, fahrenheit: 64.5 },
      { celsius: 20, fahrenheit: 68.0 },
      { celsius: 22, fahrenheit: 71.5 },
      { celsius: 24, fahrenheit: 75.0 }
    ]
  },
  safetyGuidelines: {
    title: "Safety Guidelines",
    general: [
      "Wear chemical resistant gloves and apron",
      "Use safety glasses with side shields",
      "Ensure proper ventilation",
      "Keep chemicals away from children",
      "Never mix chemicals in incorrect order"
    ],
    firstAid: {
      eyeContact: "Rinse with water for 15 minutes, seek medical attention",
      skinContact: "Wash thoroughly with soap and water",
      inhalation: "Move to fresh air, seek medical attention if symptoms persist",
      ingestion: "Do not induce vomiting, seek immediate medical attention"
    }
  }
}; 