import React from 'react';
import { FaCat } from 'react-icons/fa';
import { RiAuctionFill } from 'react-icons/ri';

const STATUS_NAMES: Record<string, string> = {
  waiting_for_players: 'Waiting for players', intro: 'Intro', themes_all: 'All themes',
  round: 'Round number', round_themes: 'Round themes', questions: 'Questions',
  question_event: 'Question event', question: 'Question', answer: 'Answer',
  question_end: 'Question end', final_themes: 'Final themes', final_bets: 'Final bets',
  final_question: 'Final question', final_answer: 'Final answer',
  final_player_answer: 'Final player answer', final_player_bet: 'Final player bet',
  game_end: 'Game over',
};

export const getStatusName = (s: string) => STATUS_NAMES[s] ?? '';

export const EventType: React.FC<{ type: string }> = ({ type }) => {
  switch (type) {
    case 'auction': return <div>Auction<br /><RiAuctionFill /></div>;
    case 'bagcat': return <div>Cat in the bag<br /><FaCat /></div>;
    case 'standard': return <div>Standard</div>;
    default: return <div />;
  }
};

export const getRoundName = (game: any) =>
  game.round?.isFinal ? 'FINAL' : 'ROUND ' + String(game.round?.number);
