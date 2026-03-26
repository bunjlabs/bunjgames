import WhirligigApi from 'services/WhirligigApi';

const API_ENDPOINT = '/api/';
const WS_ENDPOINT = '/ws/';

export const WHIRLIGIG_API = new WhirligigApi(API_ENDPOINT, WS_ENDPOINT);
