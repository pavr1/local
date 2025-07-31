// Beautiful Alert System using SweetAlert2
// This file provides modern, beautiful alerts to replace basic JavaScript alerts

console.log('🎨 ALERTS.JS: Loading SweetAlert2 utilities...');

// Alert utility functions
const Alert = {
    
    // Success alert (green)
    success: (title, text = '') => {
        return Swal.fire({
            icon: 'success',
            title: title,
            text: text,
            confirmButtonColor: '#198754',
            timer: 3000,
            showConfirmButton: false,
            toast: true,
            position: 'top-end'
        });
    },

    // Error alert (red)
    error: (title, text = '') => {
        return Swal.fire({
            icon: 'error',
            title: title,
            text: text,
            confirmButtonColor: '#dc3545'
        });
    },

    // Warning alert (yellow/orange)
    warning: (title, text = '') => {
        return Swal.fire({
            icon: 'warning',
            title: title,
            text: text,
            confirmButtonColor: '#fd7e14'
        });
    },

    // Info alert (blue)
    info: (title, text = '') => {
        return Swal.fire({
            icon: 'info',
            title: title,
            text: text,
            confirmButtonColor: '#0dcaf0'
        });
    },

    // Coming Soon alert (special styling for features)
    comingSoon: (feature) => {
        return Swal.fire({
            icon: 'info',
            title: '🚧 Coming Soon!',
            text: `${feature} is under development`,
            confirmButtonText: 'Got it!',
            confirmButtonColor: '#6f42c1',
            showClass: {
                popup: 'animate__animated animate__bounceIn'
            },
            hideClass: {
                popup: 'animate__animated animate__bounceOut'
            }
        });
    },

    // Confirmation dialog
    confirm: (title, text = '', confirmText = 'Yes') => {
        return Swal.fire({
            icon: 'question',
            title: title,
            text: text,
            showCancelButton: true,
            confirmButtonText: confirmText,
            cancelButtonText: 'Cancel',
            confirmButtonColor: '#198754',
            cancelButtonColor: '#6c757d'
        });
    },

    // Simple toast notification
    toast: (message, type = 'info') => {
        const icons = {
            success: 'success',
            error: 'error', 
            warning: 'warning',
            info: 'info'
        };

        return Swal.fire({
            icon: icons[type] || 'info',
            title: message,
            toast: true,
            position: 'top-end',
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true
        });
    },

    // Loading alert
    loading: (title = 'Loading...') => {
        return Swal.fire({
            title: title,
            allowOutsideClick: false,
            allowEscapeKey: false,
            showConfirmButton: false,
            didOpen: () => {
                Swal.showLoading();
            }
        });
    }
};

// Make Alert globally available
window.Alert = Alert;

console.log('🎨 ALERTS.JS: SweetAlert2 utilities loaded!'); 