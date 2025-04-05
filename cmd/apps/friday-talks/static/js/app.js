/**
 * Friday Talks Application JavaScript
 * Handles client-side interactions
 */

document.addEventListener('DOMContentLoaded', function() {
  // Enable Bootstrap tooltips
  const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
  tooltipTriggerList.map(function(tooltipTriggerEl) {
    return new bootstrap.Tooltip(tooltipTriggerEl);
  });
  
  // Automatically hide alerts after 5 seconds
  const alerts = document.querySelectorAll('.alert');
  alerts.forEach(function(alert) {
    setTimeout(function() {
      const bsAlert = new bootstrap.Alert(alert);
      bsAlert.close();
    }, 5000);
  });
  
  // Form validation for password confirmation
  const passwordForms = document.querySelectorAll('form:has(input[type="password"])');
  passwordForms.forEach(function(form) {
    form.addEventListener('submit', function(event) {
      const password = form.querySelector('input[name="password"]');
      const confirmPassword = form.querySelector('input[name="confirm_password"]');
      
      if (password && confirmPassword) {
        if (password.value !== confirmPassword.value) {
          event.preventDefault();
          alert('Passwords do not match');
        }
      }
    });
  });
  
  // Date helpers for preferred dates
  const dateCheckboxes = document.querySelectorAll('input[name="preferred_dates[]"]');
  if (dateCheckboxes.length > 0) {
    // "Select All" functionality if there's a select-all checkbox
    const selectAll = document.getElementById('select-all-dates');
    if (selectAll) {
      selectAll.addEventListener('change', function() {
        dateCheckboxes.forEach(function(checkbox) {
          checkbox.checked = selectAll.checked;
        });
      });
    }
    
    // Make sure at least one date is selected
    const dateForm = dateCheckboxes[0].closest('form');
    if (dateForm) {
      dateForm.addEventListener('submit', function(event) {
        const checked = Array.from(dateCheckboxes).some(checkbox => checkbox.checked);
        if (!checked) {
          event.preventDefault();
          alert('Please select at least one preferred date');
        }
      });
    }
  }
  
  // Interest level visual feedback
  const interestRadios = document.querySelectorAll('input[name="interest_level"]');
  interestRadios.forEach(function(radio) {
    radio.addEventListener('change', function() {
      // Reset all labels
      interestRadios.forEach(function(r) {
        const label = r.nextElementSibling;
        label.classList.remove('text-primary', 'fw-bold');
      });
      
      // Highlight selected label
      const selectedLabel = radio.nextElementSibling;
      selectedLabel.classList.add('text-primary', 'fw-bold');
    });
  });
});