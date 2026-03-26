import WeakestApi from 'services/WeakestApi';

const API_ENDPOINT = '/api/';
const WS_ENDPOINT = '/ws/';

export const WEAKEST_API = new WeakestApi(API_ENDPOINT, WS_ENDPOINT);
