import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';

export interface StrategyCardData {
  ticker: string;
  confidence: number;
  entry: number;
  takeProfit: number;
  stopLoss: number;
  rationale: string;
}

export interface TopPick {
  ticker: string;
  currentPrice: number;
  percentChange: number;
}

export interface ConversationTurn {
  role: 'user' | 'aura';
  text: string;
}

export interface ChatResponse {
  text: string;
  strategy?: StrategyCardData;
}

@Injectable({ providedIn: 'root' })
export class TradeService {
  private _chatInputFocus = new Subject<string>();
  chatInputFocus$ = this._chatInputFocus.asObservable();

  constructor(private http: HttpClient) {}

  getTopPicks(): Observable<TopPick[]> {
    return this.http.get<TopPick[]>('/api/top-picks');
  }

  chat(history: ConversationTurn[], message: string): Observable<ChatResponse> {
    return this.http.post<ChatResponse>('/api/chat', { history, message });
  }

  triggerChatWithTicker(ticker: string) {
    this._chatInputFocus.next(ticker);
  }
}
