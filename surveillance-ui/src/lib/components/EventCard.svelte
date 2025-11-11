<script>
  import { formatRelativeTime } from '../utils/formatters.js';
  
  export let event;
  
  function getSeverityIcon(severity) {
    const icons = {
      low: '✅',
      medium: 'ℹ️',
      high: '⚠️',
      critical: '🚨',
    };
    return icons[severity] || '📝';
  }
</script>

<div class="event-card" data-severity={event.severity}>
  <div class="event-content">
    <span class="event-icon">{getSeverityIcon(event.severity)}</span>
    <div class="event-info">
      <div class="event-header">
        <h3 class="event-title">{event.title}</h3>
        <span class="event-badge" data-severity={event.severity}>
          {event.severity}
        </span>
      </div>
      <p class="event-message">{event.message}</p>
      <p class="event-time">{formatRelativeTime(event.timestamp)}</p>
    </div>
  </div>
</div>

<style>
  .event-card {
    background: var(--surface-card);
    border-radius: var(--radius-2);
    box-shadow: var(--shadow-1);
    padding: var(--size-4);
    border-inline-start: var(--border-size-3) solid;
    transition: box-shadow var(--animation-duration-fast) var(--ease-3);
  }
  
  .event-card:hover {
    box-shadow: var(--shadow-2);
  }
  
  .event-card[data-severity="critical"] {
    border-color: var(--red-6);
  }
  
  .event-card[data-severity="high"] {
    border-color: var(--orange-6);
  }
  
  .event-card[data-severity="medium"] {
    border-color: var(--yellow-6);
  }
  
  .event-card[data-severity="low"] {
    border-color: var(--green-6);
  }
  
  .event-content {
    display: flex;
    align-items: flex-start;
    gap: var(--size-3);
  }
  
  .event-icon {
    font-size: var(--font-size-5);
    flex-shrink: 0;
  }
  
  .event-info {
    flex: 1;
    min-width: 0;
  }
  
  .event-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--size-2);
    margin-block-end: var(--size-2);
  }
  
  .event-title {
    font-size: var(--font-size-2);
    font-weight: var(--font-weight-6);
    color: var(--text-1);
    margin: 0;
  }
  
  .event-badge {
    padding: var(--size-1) var(--size-2);
    border-radius: var(--radius-round);
    font-size: var(--font-size-0);
    font-weight: var(--font-weight-5);
    text-transform: uppercase;
  }
  
  .event-badge[data-severity="critical"] {
    background: var(--red-2);
    color: var(--red-9);
  }
  
  .event-badge[data-severity="high"] {
    background: var(--orange-2);
    color: var(--orange-9);
  }
  
  .event-badge[data-severity="medium"] {
    background: var(--yellow-2);
    color: var(--yellow-9);
  }
  
  .event-badge[data-severity="low"] {
    background: var(--green-2);
    color: var(--green-9);
  }
  
  .event-message {
    font-size: var(--font-size-1);
    color: var(--text-2);
    margin-block-end: var(--size-2);
  }
  
  .event-time {
    font-size: var(--font-size-0);
    color: var(--text-3);
    margin: 0;
  }
</style>