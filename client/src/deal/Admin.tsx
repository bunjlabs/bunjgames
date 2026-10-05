import React from 'react';
import { useNavigate } from 'react-router-dom';
import { FaVolumeMute } from 'react-icons/fa';

import { Loading, Button, OvalButton, ButtonLink } from 'components/UI';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import {
  GameAdmin, AdminHeader, AdminContent, BlockContent, TextContent,
  AdminFooter, FooterItem,
} from 'components/AdminLayout';
import { CasesTable } from './CasesTable';
import { PresentsList } from './PresentsList';
import Wheel from './Wheel';
import { DEAL_API } from './api';

const STATUS_NAMES: Record<string, string> = {
  round_start: 'Round start', case_select: 'Select case', haggle: 'Haggle',
  case_swap: 'Keep or switch', reveal: 'Reveal', wheel_start: 'Wheel',
  wheel_spin: 'Wheel spinning', round_end: 'Round end', game_end: 'Game over',
};

const getStatusName = (s: string) => STATUS_NAMES[s] ?? '';

const stateContent = (game: any, onOpenCase: (i: number) => void, onSelect: (i: number) => void) => {
  switch (game.state.value) {
    case 'round_start':
      return <TextContent>Round {game.state.round}</TextContent>;
    case 'case_select':
      return <CasesTable game={game} showContent onSelect={onSelect} />;
    case 'haggle':
      return <CasesTable game={game} showContent onSelect={onOpenCase} />;
    case 'case_swap':
      return <CasesTable game={game} showContent />;
    case 'reveal':
      return <CasesTable game={game} showContent />;
    case 'wheel_start':
    case 'wheel_spin':
      return <Wheel game={game} onStop={() => {}} />;
    case 'round_end': {
      if (game.state.chosenMoney) {
        return (
          <TextContent>Money</TextContent>
        );
      }
      const idx = game.state.wheelResultIndex >= 0 ? game.state.wheelResultIndex : game.state.revealedIndex;
      return (
        <TextContent>
          {idx >= 0 && <> {idx + 1}: {game.presents[idx]?.content}</>}
        </TextContent>
      );
    }
    case 'game_end':
      return <TextContent>Deal or No Deal</TextContent>;
    default:
      return null;
  }
};

const DealAdmin: React.FC = () => {
  const game = useGame(DEAL_API);
  const [connected, setConnected] = useAuth(DEAL_API);
  const navigate = useNavigate();

  const onLogout = () => { DEAL_API.logout(); navigate('/admin'); };
  const onSoundStop = () => DEAL_API.intercom('sound_stop');

  if (!connected) return <AdminAuth api={DEAL_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const state = game.state.value;
  const onNext = () => DEAL_API.nextState(state);
  const onSelect = (i: number) => DEAL_API.selectCase(i);
  const onOpenCase = (i: number) => DEAL_API.openCase(i);
  const onRevealCase = (i: number) => DEAL_API.revealCase(i);
  const onMoney = () => DEAL_API.chooseMoney();
  const onShuffle = () => DEAL_API.shuffle();
  const onBanker = () => DEAL_API.intercom('phone_ring');

  const shuffleBtn = (
    <Button
      key="shuffle"
      onClick={onShuffle}
      style={{ fontSize: 14, padding: '4px 10px', marginRight: 24 }}
    >
      Shuffle
    </Button>
  );

  const controls: React.ReactNode[] = [];
  if (state === 'round_start' || state === 'round_end' || state === 'wheel_spin') {
    controls.push(<Button key="next" onClick={onNext}>Next</Button>);
  } else if (state === 'case_select') {
    controls.push(shuffleBtn);
  } else if (state === 'haggle') {
    controls.push(<Button key="banker" onClick={onBanker}>Banker</Button>);
  } else if (state === 'case_swap') {
    controls.push(<Button key="banker" onClick={onBanker}>Banker</Button>);
    controls.push(<Button key="keep" onClick={() => DEAL_API.keep()}>Keep</Button>);
    controls.push(<Button key="switch" onClick={() => DEAL_API.switchCase()}>Switch</Button>);
  } else if (state === 'reveal') {
    controls.push(<Button key="keep" onClick={() => DEAL_API.keepGift()}>Keep</Button>);
    controls.push(<Button key="spin" onClick={() => DEAL_API.spinWheel()}>Spin wheel</Button>);
  } else if (state === 'wheel_start') {
    const available = (game.presents as any[]).map((p, i) => ({ p, i })).filter(({ p }) => !p.given);
    controls.push(
      <select
        key="cheat"
        className="input"
        value={game.state.cheatIndex}
        onChange={(e) => DEAL_API.setCheat(parseInt(e.target.value))}
        style={{ maxWidth: 200 }}
      >
        <option value={-1}>No cheat</option>
        {available.map(({ p, i }) => (
          <option key={i} value={i}>{p.content}</option>
        ))}
      </select>,
    );
    controls.push(<Button key="spin" onClick={() => DEAL_API.spin()}>Spin</Button>);
  }

  return (
    <GameAdmin>
      <AdminHeader gameName="Deal or No Deal" token={game.token} stateName={getStatusName(state)}>
        <OvalButton onClick={onSoundStop}><FaVolumeMute /></OvalButton>
        <ButtonLink to="/admin">Home</ButtonLink>
        <ButtonLink to="/deal/view">View</ButtonLink>
        <Button onClick={onLogout}>Logout</Button>
      </AdminHeader>
      <AdminContent rightPanel={<div style={{ padding: 8, flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}><PresentsList game={game} showDescription onClick={(state === 'haggle' || state === 'case_swap') ? onRevealCase : undefined} onMoney={(state === 'haggle' || state === 'case_swap') ? onMoney : undefined} /></div>}>
        <BlockContent>
          {stateContent(game, onOpenCase, onSelect)}
        </BlockContent>
      </AdminContent>
      <AdminFooter>
        <FooterItem style={{ fontSize: 30 }}>
          {state === 'wheel_spin' && game.state.wheelResultIndex >= 0
            ? game.presents[game.state.wheelResultIndex]?.content
            : `Round ${game.state.round}`}
        </FooterItem>
        <FooterItem>{controls}</FooterItem>
      </AdminFooter>
    </GameAdmin>
  );
};

export default DealAdmin;
