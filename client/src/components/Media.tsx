import React, { useEffect, useRef } from 'react';
import { Howl } from 'howler';

export const HowlWrapper = (src: string, loop = false, volume = 1.0) =>
  new Howl({ src: [src], loop, volume, preload: false });

const isAbsoluteUrl = (url: string) => /^https?:\/\//.test(url);

const getMediaUrl = (game: { token: string }, url: string) =>
  isAbsoluteUrl(url)
    ? url
    : `/media/${game.token}${url.startsWith('/') ? '' : '/'}${url}`;

export const ImagePlayer: React.FC<{ game: { token: string }; url: string; style?: React.CSSProperties }> = ({ game, url, style }) => (
  <img src={getMediaUrl(game, url)} alt="Missing" style={{ maxWidth: '100%', maxHeight: '100%', objectFit: 'contain', display: 'block', ...style }} />
);

export const AudioPlayer: React.FC<{
  game: { token: string };
  url: string;
  controls?: boolean;
  playing?: boolean;
}> = ({ game, url, controls, playing }) => {
  const audioUrl = getMediaUrl(game, url);
  const audioRef = useRef<HTMLAudioElement>(null);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const handleCanPlay = () => {
      if (playing) audio.play().catch(() => {});
    };

    audio.addEventListener('canplay', handleCanPlay);
    if (playing && audio.readyState >= 3) audio.play().catch(() => {});

    return () => audio.removeEventListener('canplay', handleCanPlay);
  }, [playing, audioUrl]);

  return (
    <audio ref={audioRef} src={audioUrl} controls={controls} style={{ width: '80%' }} preload="metadata" />
  );
};

export const VideoPlayer: React.FC<{
  game: { token: string };
  url: string;
  controls?: boolean;
  playing?: boolean;
  style?: React.CSSProperties;
}> = ({ game, url, controls, playing, style }) => {
  const videoUrl = getMediaUrl(game, url);
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handleCanPlay = () => {
      if (playing) video.play().catch(() => {});
    };

    video.addEventListener('canplay', handleCanPlay);
    if (playing && video.readyState >= 3) video.play().catch(() => {});

    return () => video.removeEventListener('canplay', handleCanPlay);
  }, [playing, videoUrl]);

  return (
    <video
      ref={videoRef}
      src={videoUrl}
      controls={controls}
      style={{ maxWidth: '100%', maxHeight: '100%', objectFit: 'contain', display: 'block', ...style }}
      preload="metadata"
      playsInline
    />
  );
};
