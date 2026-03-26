export type SubscriberCallback<T> = (payload: T, id: string) => void;

export default class Subscriber<T = any> {
  private nextId = 0;
  private subscriptions: Record<string, SubscriberCallback<T>> = {};
  private last: T | undefined;
  private readonly isState: boolean;

  constructor(isState = true) {
    this.isState = isState;
  }

  subscribe(callback: SubscriberCallback<T>): number {
    const id = this.nextId++;
    this.subscriptions[id] = callback;
    if (this.isState && this.last !== undefined) callback(this.last, String(id));
    return id;
  }

  unsubscribe(id: number) {
    delete this.subscriptions[id];
  }

  fire(payload: T) {
    this.last = payload;
    for (const [id, cb] of Object.entries(this.subscriptions)) {
      cb(payload, id);
    }
  }
}
