/** Svelte 适配层：只负责把框架无关的 Session 订阅接入 Svelte 响应式，不持有业务状态。 */
import { createSubscriber } from 'svelte/reactivity';
import type { Session, Snapshot } from '../core/session';

export class SessionStore {
  readonly session: Session;
  #subscribe: () => void;
  constructor(session: Session) {
    this.session = session;
    this.#subscribe = createSubscriber((update) => session.subscribe(update));
  }
  get snap(): Snapshot {
    this.#subscribe();
    return this.session.getSnapshot();
  }
}
