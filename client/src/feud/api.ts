import FeudApi from 'services/FeudApi';

const API_ENDPOINT = '/api/';
const WS_ENDPOINT = '/ws/';

export const FEUD_API = new FeudApi(API_ENDPOINT, WS_ENDPOINT);
