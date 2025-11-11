import { writable } from 'svelte/store';
import { api } from '../utils/api.js';

function createEventsStore() {
  const { subscribe, set, update } = writable({
    events: [],
    loading: false,
    error: null,
  });

  return {
    subscribe,
    
    fetch: async function(limit = 20) {
      console.log('📢 Fetching events...');
      update(state => ({ ...state, loading: true, error: null }));
      
      try {
        const response = await api.getEvents(limit);
        console.log('✅ Events response:', response);
        
        if (response && response.success && response.data) {
          update(state => ({
            ...state,
            events: response.data,
            loading: false,
          }));
          console.log('✅ Events updated');
        } else {
          throw new Error('Invalid response');
        }
      } catch (error) {
        console.error('❌ Error fetching events:', error);
        update(state => ({
          ...state,
          error: error.message,
          loading: false,
        }));
      }
    },

    reset: function() {
      set({
        events: [],
        loading: false,
        error: null,
      });
    }
  };
}

export const eventsStore = createEventsStore();