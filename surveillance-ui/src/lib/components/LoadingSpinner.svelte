<script>
  export let size = 'md';
</script>

<div class="spinner-container">
  <div class="spinner" data-size={size}></div>
</div>

<style>
  .spinner-container {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--size-8);
  }
  
  .spinner {
    border: var(--border-size-3) solid var(--surface-3);
    border-top-color: var(--primary);
    border-radius: var(--radius-round);
    animation: spin 1s linear infinite;
  }
  
  .spinner[data-size="sm"] {
    width: var(--size-4);
    height: var(--size-4);
  }
  
  .spinner[data-size="md"] {
    width: var(--size-8);
    height: var(--size-8);
  }
  
  .spinner[data-size="lg"] {
    width: var(--size-12);
    height: var(--size-12);
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>