import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, Subject } from 'rxjs';

export interface ModelInfo {
  id: string;
  name: string;
  provider: 'google' | 'ollama';
}

export interface StrategyCardData {
  ticker: string;
  confidence: number;
  entry: number;
  takeProfit: number;
  stopLoss: number;
  rationale: string;
  // Technical Analysis
  rsi?: number;
  macd?: number;
  macdSignal?: number;
  macdHistogram?: number;
  ma20?: number;
  ma50?: number;
  atr?: number;
  stopLossJustification?: string;
  takeProfitJustification?: string;
}

export interface TopPick {
  ticker: string;
  currentPrice: number;
  percentChange: number;
  volume?: number;
  rsi?: number;
  signal?: string;
}

export interface ConversationTurn {
  role: 'user' | 'aura';
  text: string;
}

export interface ChatResponse {
  text: string;
  strategy?: StrategyCardData;
}

export interface ScreenerResult {
  ticker: string;
  currentPrice: number;
  percentChange: number;
  volume: number;
  rsi: number;
  ma20: number;
  ma50: number;
  macd: number;
  macdSignal: number;
  signal: 'bullish' | 'bearish' | 'neutral';
}

export interface ScreenerFilters {
  minPrice?: number;
  maxPrice?: number;
  minVolume?: number;
  minRSI?: number;
  maxRSI?: number;
  signal?: string;
}

export interface Alert {
  id: number;
  ticker: string;
  condition: 'above' | 'below';
  targetPrice: number;
  isActive: boolean;
  createdAt: string;
}

export interface CreateAlertRequest {
  ticker: string;
  condition: 'above' | 'below';
  targetPrice: number;
}

@Injectable({ providedIn: 'root' })
export class TradeService {
  private _chatInputFocus = new Subject<string>();
  chatInputFocus$ = this._chatInputFocus.asObservable();

  private _alertsChanged = new BehaviorSubject<void>(undefined);
  alertsChanged$ = this._alertsChanged.asObservable();

  /** Currently selected AI model ID. Defaults to gemini-2.5-pro. */
  private _selectedModel = new BehaviorSubject<string>('gemini-2.5-pro');
  selectedModel$ = this._selectedModel.asObservable();

  get selectedModel(): string { return this._selectedModel.getValue(); }
  setModel(id: string) { this._selectedModel.next(id); }

  notifyAlertsChanged() {
    this._alertsChanged.next();
  }

  constructor(private http: HttpClient) {}

  getModels(): Observable<ModelInfo[]> {
    return this.http.get<ModelInfo[]>('/api/models');
  }

  getTopPicks(): Observable<TopPick[]> {
    return this.http.get<TopPick[]>('/api/top-picks');
  }

  chat(history: ConversationTurn[], message: string): Observable<ChatResponse> {
    return this.http.post<ChatResponse>('/api/chat', { history, message, model: this.selectedModel });
  }

  analyze(ticker: string): Observable<StrategyCardData> {
    return this.http.post<StrategyCardData>('/api/analyze', { ticker, model: this.selectedModel });
  }

  screen(filters: ScreenerFilters): Observable<ScreenerResult[]> {
    const params: Record<string, string> = {};
    if (filters.minPrice != null) params['minPrice'] = String(filters.minPrice);
    if (filters.maxPrice != null) params['maxPrice'] = String(filters.maxPrice);
    if (filters.minVolume != null) params['minVolume'] = String(filters.minVolume);
    if (filters.minRSI != null) params['minRSI'] = String(filters.minRSI);
    if (filters.maxRSI != null) params['maxRSI'] = String(filters.maxRSI);
    if (filters.signal) params['signal'] = filters.signal;
    return this.http.get<ScreenerResult[]>('/api/screener', { params });
  }

  getAlerts(): Observable<Alert[]> {
    return this.http.get<Alert[]>('/api/alerts');
  }

  createAlert(req: CreateAlertRequest): Observable<{ id: number; message: string }> {
    return this.http.post<{ id: number; message: string }>('/api/alerts', req);
  }

  deleteAlert(id: number): Observable<void> {
    return this.http.delete<void>(`/api/alerts?id=${id}`);
  }

  triggerChatWithTicker(ticker: string) {
    this._chatInputFocus.next(ticker);
  }

  clearChatTrigger() {
    // No longer needed with Subject
  }
}
