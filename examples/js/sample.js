/**
 * This is a sample JavaScript module that demonstrates various function types
 * and JSDoc commenting styles.
 *
 * @module SampleModule
 */

/**
 * Adds two numbers together
 *
 * @param {number} a - First number
 * @param {number} b - Second number
 * @returns {number} The sum of a and b
 */
function add(a, b) {
  return a + b;
}

/**
 * Subtracts one number from another
 *
 * @param {number} a - Number to subtract from
 * @param {number} b - Number to subtract
 * @returns {number} The difference of a and b
 */
const subtract = function (a, b) {
  return a - b;
};

/**
 * Multiplies two numbers
 *
 * @param {number} a - First number
 * @param {number} b - Second number
 * @returns {number} The product of a and b
 */
const multiply = (a, b) => {
  return a * b;
};

// This is a regular comment, not a JSDoc comment
function divide(a, b) {
  if (b === 0) {
    throw new Error("Cannot divide by zero");
  }
  return a / b;
}

/**
 * A sample class that demonstrates method docstrings
 */
class Calculator {
  /**
   * Creates a new Calculator
   *
   * @param {number} initialValue - The starting value
   */
  constructor(initialValue = 0) {
    this.value = initialValue;
  }

  /**
   * Adds a number to the current value
   *
   * @param {number} num - Number to add
   * @returns {number} The new value
   */
  add(num) {
    this.value += num;
    return this.value;
  }

  /**
   * Subtracts a number from the current value
   *
   * @param {number} num - Number to subtract
   * @returns {number} The new value
   */
  subtract(num) {
    this.value -= num;
    return this.value;
  }
}

// Object with methods
const mathUtils = {
  /**
   * Calculate the square of a number
   *
   * @param {number} x - Input number
   * @returns {number} The square of x
   */
  square(x) {
    return x * x;
  },

  /**
   * Calculate the cube of a number
   *
   * @param {number} x - Input number
   * @returns {number} The cube of x
   */
  cube(x) {
    return x * x * x;
  },
};
