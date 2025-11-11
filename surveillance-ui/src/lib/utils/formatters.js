import { formatDistanceToNow, format } from 'date-fns';

export function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

export function formatDuration(seconds) {
  if (!seconds) return '0s';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}

export function formatRelativeTime(date) {
  if (!date) return 'Never';
  return formatDistanceToNow(new Date(date), { addSuffix: true });
}

export function formatDateTime(date) {
  if (!date) return 'N/A';
  return format(new Date(date), 'MMM dd, HH:mm:ss');
}

export function getSeverityColor(severity) {
  const colors = {
    low: 'success',
    medium: 'warning', 
    high: 'danger',
    critical: 'critical',
  };
  return colors[severity] || 'info';
}

export function getSeverityIcon(severity) {
  const icons = {
    low: '✅',
    medium: 'ℹ️',
    high: '⚠️',
    critical: '🚨',
  };
  return icons[severity] || '📝';
}