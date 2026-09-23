import DealApi from 'services/DealApi';

const API_ENDPOINT = '/api/';
const WS_ENDPOINT = '/ws/';

export const DEAL_API = new DealApi(API_ENDPOINT, WS_ENDPOINT);
