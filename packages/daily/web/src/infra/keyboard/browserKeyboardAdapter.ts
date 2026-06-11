/** Binds global keyboard events to a callback. */
export class BrowserKeyboardAdapter {
  private bound = false;

  private listener: ((event: KeyboardEvent) => void) | null = null;

  private readonly listenerOptions: AddEventListenerOptions = {
    capture: true,
  };

  start(onKeydown: (event: KeyboardEvent) => void): void {
    if (this.bound) return;
    this.listener = onKeydown;
    window.addEventListener('keydown', onKeydown, this.listenerOptions);
    this.bound = true;
  }

  stop(): void {
    if (!this.bound || !this.listener) return;
    window.removeEventListener('keydown', this.listener, this.listenerOptions);
    this.listener = null;
    this.bound = false;
  }
}
