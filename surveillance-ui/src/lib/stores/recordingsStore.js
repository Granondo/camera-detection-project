import { writable } from 'svelte/store';
import { api } from '../utils/api.js';

function createRecordingsStore() {
  const { subscribe, set, update } = writable({
    recordings: [],
    loading: false,
    error: null,
  });

  return {
    subscribe,
    
    fetch: async function(limit = 20) {
      console.log('📼 Fetching recordings...');
      update(state => ({ ...state, loading: true, error: null }));
      
      try {
        const response = await api.getRecordings(limit);
        console.log('✅ Recordings response:', response);
        
        if (response && response.success && response.data) {
          update(state => ({
            ...state,
            recordings: response.data,
            loading: false,
          }));
          console.log('✅ Recordings updated');
        } else {
          throw new Error('Invalid response');
        }
      } catch (error) {
        console.error('❌ Error fetching recordings:', error);
        update(state => ({
          ...state,
          error: error.message,
          loading: false,
        }));
      }
    },

    reset: function() {
      set({
        recordings: [],
        loading: false,
        error: null,
      });
    }
  };
}

export const recordingsStore = createRecordingsStore();