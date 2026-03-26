import { useEffect, useState } from 'react';
import type GameApi from 'services/GameApi';
import type { GameState } from 'services/GameApi';

export const useGame = (
  api: GameApi,
  onState: (game: GameState) => void = () => {},
  onIntercom: (message: string) => void = () => {},
): GameState | undefined => {
  const [game, setGame] = useState<GameState>();

  useEffect(() => {
    const gameId = api.getGameSubscriber().subscribe(setGame);
    const stateId = api.getStateSubscriber().subscribe(onState);
    const intercomId = api.getIntercomSubscriber().subscribe(onIntercom);
    return () => {
      api.getGameSubscriber().unsubscribe(gameId);
      api.getStateSubscriber().unsubscribe(stateId);
      api.getIntercomSubscriber().unsubscribe(intercomId);
    };
  }, [api, onIntercom, onState]);

  return game;
};

export const useAuth = (api: GameApi): [boolean | undefined, React.Dispatch<React.SetStateAction<boolean | undefined>>] => {
  const [connected, setConnected] = useState<boolean>();
  useEffect(() => setConnected(api.isConnected()), [api]);
  return [connected, setConnected];
};

export const useTimer = (
  api: { calcTime: () => number; next_state?: (s: string) => void; nextState?: (s: string) => void },
  onTimerEnd?: () => void,
): number => {
  const [time, setTime] = useState(api.calcTime());

  useEffect(() => {
    let timer: ReturnType<typeof setInterval>;
    if (time <= 0) {
      if (onTimerEnd) onTimerEnd();
    } else {
      timer = setInterval(() => setTime(api.calcTime()), 1000);
    }
    return () => timer && clearInterval(timer);
  });

  return time;
};

export const calcStateName = (state: string): string => {
  const value = state.replaceAll('_', ' ');
  return value.charAt(0).toUpperCase() + value.slice(1);
};
