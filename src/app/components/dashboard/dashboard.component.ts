import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TopPicksComponent } from '../top-picks/top-picks.component';
import { ChatComponent } from '../chat/chat.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, TopPicksComponent, ChatComponent],
  template: `
    <div class="flex flex-col h-full bg-slate-950 font-sans text-slate-200">
      <app-top-picks></app-top-picks>
      <main class="flex-1 overflow-hidden">
        <app-chat></app-chat>
      </main>
    </div>
  `,
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent {

}
