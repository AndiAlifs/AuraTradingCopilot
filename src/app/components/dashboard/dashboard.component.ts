import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TopPicksComponent } from '../top-picks/top-picks.component';
import { ChatComponent } from '../chat/chat.component';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink, TopPicksComponent, ChatComponent],
  template: `
    <div class="flex flex-col h-screen bg-slate-950 font-sans text-slate-200">

      <!-- Top nav -->
      <header class="flex items-center justify-between px-4 py-2 border-b border-auraBorder bg-auraPanel shrink-0">
        <div class="flex items-center gap-1 text-lg font-black tracking-tight">
          <span class="text-white">Aura</span>
          <span class="text-auraNeon"> Copilot</span>
        </div>
        <nav class="flex items-center gap-3 text-sm">
          <a routerLink="/screener"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L13 13.414V19a1 1 0 01-.553.894l-4 2A1 1 0 017 21v-7.586L3.293 6.707A1 1 0 013 6V4z" />
            </svg>
            Screener
          </a>
          <a routerLink="/alerts"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            Alerts
          </a>
          <ng-container *ngIf="email">
            <div class="h-4 w-px bg-slate-700"></div>
            <span class="text-xs text-slate-500 hidden sm:block">{{ email }}</span>
            <button (click)="logout()"
              class="px-3 py-1.5 text-xs text-slate-500 hover:text-slate-300 transition-colors rounded-lg hover:bg-slate-800">
              Sign out
            </button>
          </ng-container>
          <ng-container *ngIf="!email">
            <div class="h-4 w-px bg-slate-700"></div>
            <a routerLink="/login"
              class="px-3 py-1.5 text-xs text-auraNeon hover:text-cyan-300 transition-colors rounded-lg hover:bg-slate-800">
              Sign in
            </a>
          </ng-container>
        </nav>
      </header>

      <!-- Ticker bar -->
      <app-top-picks></app-top-picks>

      <!-- Chat area -->
      <main class="flex-1 overflow-hidden">
        <app-chat></app-chat>
      </main>
    </div>
  `,
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent {
  constructor(private auth: AuthService) {}

  get email() { return this.auth.email; }

  logout() { this.auth.logout(); }
}
