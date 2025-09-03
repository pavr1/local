/**
 * Simple Response Status Validation
 * 
 * This file provides basic status code validation for API responses
 * to work with the Go backend response structure:
 * {
 *   "code": 200,           // HTTP status code (200-299 = success)
 *   "message": "Success",  // Optional message string
 *   "data": {...}          // The actual data payload
 * }
 */

/**
 * Check if response code indicates success
 * @param {Object} result - API response object
 * @returns {boolean} - True if response code is in success range (200-299)
 */
function isSuccessfulResponse(result) {
    return result.code >= 200 && result.code < 300;
}

// Export function for global use
window.isSuccessfulResponse = isSuccessfulResponse;

console.log('🔧 Response status validation loaded successfully');
