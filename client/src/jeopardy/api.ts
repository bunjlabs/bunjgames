import JeopardyApi from 'services/JeopardyApi';

const API_ENDPOINT = '/api/';
const WS_ENDPOINT = '/ws/';

export const JEOPARDY_API = new JeopardyApi(API_ENDPOINT, WS_ENDPOINT);
