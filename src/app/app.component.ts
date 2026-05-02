import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TopPicksComponent } from './components/top-picks/top-picks.component';
import { ChatComponent } from './components/chat/chat.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, TopPicksComponent, ChatComponent],
  template: `
    <div class="flex flex-col h-screen bg-slate-950 font-sans text-slate-200">
      <app-top-picks></app-top-picks>
      <main class="flex-1 overflow-hidden">
        <app-chat></app-chat>
      </main>
    </div>
  `
})
export class AppComponent {
  title = 'aura-trade';
}
