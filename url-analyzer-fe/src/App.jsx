import React, { useState, useEffect, useRef } from 'react';
import { 
  Globe, 
  Play, 
  Copy, 
  Check, 
  RefreshCw, 
  AlertCircle, 
  Sparkles, 
  ExternalLink,
  Link as LinkIcon,
  Code,
  FileText,
  Search,
  ChevronDown,
  ChevronUp,
  ShieldCheck,
  ShieldAlert,
  Heading,
  AlertOctagon
} from 'lucide-react';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

if (!API_BASE_URL) {
  throw new Error("Missing environment variable: VITE_API_BASE_URL must be defined at build time.");
}

export default function App() {
  const [url, setUrl] = useState('https://github.com');
  const [loading, setLoading] = useState(false);
  
  // Timeout settings fetched from API
  const [defaultTimeout, setDefaultTimeout] = useState(20);
  const [timeLeft, setTimeLeft] = useState(20);

  const [responseJson, setResponseJson] = useState(null);
  const [error, setError] = useState(null);
  const [copied, setCopied] = useState(false);

  const timerRef = useRef(null);

  // Fetch timeout configuration on component mount
  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/v1/config`);
        if (res.ok) {
          const data = await res.json();
          if (data && typeof data.scrape_timeout_seconds === 'number') {
            setDefaultTimeout(data.scrape_timeout_seconds);
            setTimeLeft(data.scrape_timeout_seconds);
          }
        }
      } catch (err) {
        console.warn('Could not fetch config from backend, using default 20s fallback:', err);
      }
    };

    fetchConfig();
  }, []);

  const handleScrape = async (e) => {
    e.preventDefault();
    
    const trimmedUrl = url.trim();
    if (!trimmedUrl) {
      setResponseJson(null); // Clear previous output on validation error
      setError({
        code: 'VALIDATION_ERROR',
        message: 'Please enter a valid URL.',
      });
      return;
    }

    setLoading(true);
    setError(null);
    setResponseJson(null); // Clear previous output when starting a new scrape
    
    // Reset timer
    const startDuration = defaultTimeout;
    setTimeLeft(startDuration);

    if (timerRef.current) clearInterval(timerRef.current);
    
    timerRef.current = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timerRef.current);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    try {
      const payload = {
        urls: [
          {
            url: trimmedUrl
          }
        ]
      };

      const res = await fetch(`${API_BASE_URL}/api/v1/scrape`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      const data = await res.json();

      // 1. Check for top-level batch/API error
      if (!res.ok || data.error) {
        setResponseJson(null); // Clear previous result on error
        setError(data.error || { code: `HTTP_${res.status}`, message: res.statusText });
        return;
      }

      // 2. Check for item-level URL processing errors
      if (Array.isArray(data.data) && data.data.length > 0) {
        const itemWithError = data.data.find((item) => item.error);
        if (itemWithError) {
          setResponseJson(null); // Clear previous result on item error
          setError({
            url: itemWithError.url,
            ...itemWithError.error
          });
          return;
        }
      }

      // Successful analysis result
      setResponseJson(data);

    } catch (err) {
      setResponseJson(null); // Clear previous result on connection error
      setError({
        code: 'CONNECTION_ERROR',
        message: err.message || 'Failed to connect to the API server.',
      });
    } finally {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      setLoading(false);
    }
  };

  const handleCopyRaw = () => {
    if (responseJson) {
      navigator.clipboard.writeText(JSON.stringify(responseJson, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const progressPercentage = defaultTimeout > 0 
    ? ((defaultTimeout - timeLeft) / defaultTimeout) * 100 
    : 100;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      {/* Background Glows */}
      <div className="fixed top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[300px] bg-indigo-600/15 blur-[120px] pointer-events-none rounded-full" />
      <div className="fixed bottom-10 right-10 w-80 h-80 bg-cyan-600/10 blur-[100px] pointer-events-none rounded-full" />

      {/* Main Container */}
      <div className="w-full max-w-5xl relative z-10 space-y-8">
        
        {/* Header */}
        <div className="text-center space-y-3">
          <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-slate-900 border border-slate-800 text-xs text-indigo-400 font-medium">
            <Sparkles className="w-3.5 h-3.5" /> lucytech-url-analyzer-fe
          </div>
          <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
            Lucytech Assignment
          </h1>
          <p className="text-slate-400 text-sm sm:text-base max-w-lg mx-auto">
            Enter a URL below to process and inspect page metadata.
          </p>
        </div>

        {/* Input Card */}
        <div className="bg-slate-900/90 border border-slate-800 backdrop-blur-xl rounded-2xl p-6 sm:p-8 shadow-2xl space-y-6">
          <form onSubmit={handleScrape} className="space-y-4">
            <div className="flex flex-col space-y-2">
              <label className="text-xs font-semibold uppercase tracking-wider text-slate-400 flex justify-between items-center">
                <span>Target Web Page URL</span>
              </label>

              <div className="relative flex items-center">
                <Globe className="absolute left-4 w-5 h-5 text-slate-500 pointer-events-none" />
                <input
                  type="url"
                  required
                  placeholder="https://www.scrapethissite.com/pages/"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  disabled={loading}
                  className="w-full bg-slate-950 border border-slate-800 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 text-slate-100 placeholder-slate-600 rounded-xl pl-12 pr-36 py-4 text-sm sm:text-base transition-all outline-none"
                />
                <button
                  type="submit"
                  disabled={loading}
                  className="absolute right-2 top-2 bottom-2 px-5 rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-medium text-sm transition-all flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? (
                    <>
                      <RefreshCw className="w-4 h-4 animate-spin" />
                      <span>Processing</span>
                    </>
                  ) : (
                    <>
                      <Play className="w-4 h-4 fill-current" />
                      <span>Process</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          </form>

          {/* Loading Progress & Countdown Bar */}
          {loading && (
            <div className="pt-4 border-t border-slate-800 space-y-3">
              <div className="flex items-center justify-between text-xs font-mono">
                <span className="text-slate-400 flex items-center gap-2">
                  <RefreshCw className="w-3.5 h-3.5 animate-spin text-indigo-400" />
                  Fetching response from server...
                </span>
                <span className="text-indigo-400 font-bold bg-indigo-500/10 px-2.5 py-1 rounded-md border border-indigo-500/20">
                  {timeLeft}s remaining
                </span>
              </div>

              <div className="w-full bg-slate-950 h-2.5 rounded-full overflow-hidden border border-slate-800">
                <div
                  className="bg-gradient-to-r from-indigo-500 to-cyan-400 h-full transition-all duration-1000 ease-linear rounded-full"
                  style={{ width: `${progressPercentage}%` }}
                />
              </div>
            </div>
          )}

          {/* Structured Error Alert Card */}
          {error && (
            <div className="p-4 rounded-xl bg-red-950/40 border border-red-500/30 text-red-300 space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 font-semibold text-red-400 text-sm">
                  <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />
                  <span>Processing Error</span>
                </div>
                {error.code && (
                  <span className="text-[10px] font-mono uppercase bg-red-500/20 text-red-300 px-2 py-0.5 rounded border border-red-500/30">
                    {error.code}
                  </span>
                )}
              </div>

              {error.message && (
                <p className="text-xs text-slate-300 pl-6 leading-relaxed">
                  {error.message}
                </p>
              )}

              {error.url && (
                <div className="pl-6 text-[11px] font-mono text-slate-400 truncate">
                  <span className="text-slate-500">Target URL: </span>{error.url}
                </div>
              )}

              {Array.isArray(error.details) && error.details.length > 0 && (
                <div className="pl-6 pt-1">
                  <ul className="list-disc list-inside text-xs text-slate-400 space-y-1 bg-slate-950/50 p-2.5 rounded-lg border border-red-500/10">
                    {error.details.map((detail, idx) => (
                      <li key={idx}>{detail}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Custom Visual Analysis View */}
        {responseJson && (
          <div className="space-y-4">
            <div className="flex items-center justify-between px-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Analysis Output</span>
              <button
                onClick={handleCopyRaw}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium border border-slate-800 transition-all"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copied ? 'Copied Raw JSON' : 'Copy Raw JSON'}</span>
              </button>
            </div>

            <AnalysisResultView responseData={responseJson} />
          </div>
        )}

      </div>
    </div>
  );
}

// Custom UI Renderer for Scrape Analysis Data
function AnalysisResultView({ responseData }) {
  if (!responseData || !Array.isArray(responseData.data) || responseData.data.length === 0) {
    return null;
  }

  const result = responseData.data[0] || {};
  const analysis = result.analysis || {};

  return (
    <div className="w-full space-y-6 text-slate-100 font-sans">
      {/* 1. Overview Header & Meta Info */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 backdrop-blur-xl shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-slate-800">
          <div className="flex items-center gap-3 min-w-0">
            <div className="p-2.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 shrink-0">
              <Globe className="w-6 h-6" />
            </div>
            <div className="min-w-0">
              <span className="text-xs uppercase tracking-wider text-slate-400 font-semibold">Target URL</span>
              <h2 className="text-base sm:text-lg font-bold text-white truncate">
                {result.url || analysis.url || 'N/A'}
              </h2>
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <span className={`px-3 py-1 rounded-full text-xs font-semibold border flex items-center gap-1.5 ${
              responseData.status === 200 
                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
                : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
            }`}>
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              Status: {responseData.status || 200}
            </span>
          </div>
        </div>

        {/* Quick Data Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 pt-2">
          {/* Page Title Card */}
          <div className="sm:col-span-2 bg-slate-950/60 border border-slate-800/80 rounded-xl p-4 flex flex-col justify-between">
            <div className="flex items-center gap-2 text-xs font-semibold text-slate-400 mb-1">
              <FileText className="w-4 h-4 text-indigo-400" />
              <span>Page Title</span>
            </div>
            <p className="text-sm font-medium text-slate-200 line-clamp-2">
              {analysis.title || 'No Title Found'}
            </p>
          </div>

          {/* HTML Version */}
          <div className="bg-slate-950/60 border border-slate-800/80 rounded-xl p-4 flex flex-col justify-between">
            <div className="flex items-center gap-2 text-xs font-semibold text-slate-400 mb-1">
              <Code className="w-4 h-4 text-cyan-400" />
              <span>HTML Version</span>
            </div>
            <p className="text-base font-bold text-slate-100">
              {analysis.html_version || 'N/A'}
            </p>
          </div>

          {/* Login Form Status */}
          <div className="bg-slate-950/60 border border-slate-800/80 rounded-xl p-4 flex flex-col justify-between">
            <div className="flex items-center gap-2 text-xs font-semibold text-slate-400 mb-1">
              <span>Login Form</span>
            </div>
            <div className="flex items-center gap-2">
              {analysis.has_login_form ? (
                <span className="text-amber-400 font-semibold text-sm flex items-center gap-1.5">
                  <ShieldAlert className="w-4 h-4" /> Detected
                </span>
              ) : (
                <span className="text-emerald-400 font-semibold text-sm flex items-center gap-1.5">
                  <ShieldCheck className="w-4 h-4" /> Not Found
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Headings Breakdown */}
        {analysis.headings && (
          <div className="bg-slate-950/40 border border-slate-800/60 rounded-xl p-4 flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2 text-xs font-semibold text-slate-400 mr-2">
              <Heading className="w-4 h-4 text-indigo-400" />
              <span>Headings Structure:</span>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {Object.entries(analysis.headings).map(([tag, count]) => (
                <span key={tag} className="px-2.5 py-1 rounded-md bg-slate-800 border border-slate-700 text-xs font-mono text-indigo-300">
                  <strong className="text-slate-100 uppercase">{tag}:</strong> {count}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* 2. Interactive Link Cards with Search & Filters */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <LinkListCard
          title="Internal Links"
          links={analysis.internal_links || []}
          icon={<LinkIcon className="w-4 h-4 text-indigo-400" />}
          badgeColor="bg-indigo-500/10 text-indigo-400 border-indigo-500/20"
        />
        <LinkListCard
          title="External Links"
          links={analysis.external_links || []}
          icon={<ExternalLink className="w-4 h-4 text-cyan-400" />}
          badgeColor="bg-cyan-500/10 text-cyan-400 border-cyan-500/20"
        />
        <LinkListCard
          title="Inaccessible Links"
          links={analysis.inaccessible_links || []}
          icon={<AlertOctagon className="w-4 h-4 text-rose-400" />}
          badgeColor="bg-rose-500/10 text-rose-400 border-rose-500/20"
          isDanger={true}
        />
      </div>
    </div>
  );
}

// Collapsible & Filterable Link Section
function LinkListCard({ title, links, icon, badgeColor, isDanger = false }) {
  const [isOpen, setIsOpen] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [copiedIndex, setCopiedIndex] = useState(null);

  const filteredLinks = links.filter((item) => {
    const targetUrl = typeof item === 'string' ? item : item?.url || '';
    return targetUrl.toLowerCase().includes(searchTerm.toLowerCase());
  });

  const handleCopyLink = (url, index) => {
    navigator.clipboard.writeText(url);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 1500);
  };

  return (
    <div className={`bg-slate-900/90 border rounded-2xl overflow-hidden backdrop-blur-xl shadow-2xl flex flex-col ${
      isDanger ? 'border-rose-900/40' : 'border-slate-800'
    }`}>
      {/* Header Bar */}
      <div 
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between px-5 py-4 bg-slate-950/80 border-b border-slate-800 cursor-pointer hover:bg-slate-900/60 transition-colors"
      >
        <div className="flex items-center gap-2.5">
          {icon}
          <h3 className="font-semibold text-slate-200 text-sm">{title}</h3>
          <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold border ${badgeColor}`}>
            {links.length}
          </span>
        </div>
        <button type="button" className="text-slate-400 hover:text-slate-200 transition-colors">
          {isOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </button>
      </div>

      {isOpen && (
        <div className="p-4 space-y-3 flex-1 flex flex-col">
          {/* Search Box */}
          {links.length > 5 && (
            <div className="relative">
              <Search className="absolute left-3 top-2.5 w-3.5 h-3.5 text-slate-500" />
              <input
                type="text"
                placeholder={`Search ${links.length} links...`}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-colors"
              />
            </div>
          )}

          {/* Links Scrollable Area */}
          <div className="max-h-64 overflow-y-auto space-y-1.5 pr-1">
            {links.length === 0 ? (
              <p className="text-xs text-slate-500 text-center py-4">0 inaccessible links detected.</p>
            ) : filteredLinks.length === 0 ? (
              <p className="text-xs text-slate-500 text-center py-4">No matching links found.</p>
            ) : (
              filteredLinks.map((item, idx) => {
                const linkUrl = typeof item === 'string' ? item : item?.url || '';
                return (
                  <div
                    key={idx}
                    className={`group flex items-center justify-between gap-2 p-2 rounded-lg border transition-all text-xs ${
                      isDanger 
                        ? 'bg-rose-950/20 hover:bg-rose-950/40 border-rose-900/30' 
                        : 'bg-slate-950/40 hover:bg-slate-800/50 border-slate-800/40'
                    }`}
                  >
                    <a
                      href={linkUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={`truncate flex-1 font-mono text-[11px] transition-colors ${
                        isDanger 
                          ? 'text-rose-300 hover:text-rose-200' 
                          : 'text-slate-300 hover:text-indigo-400'
                      }`}
                      title={linkUrl}
                    >
                      {linkUrl}
                    </a>

                    <button
                      type="button"
                      onClick={() => handleCopyLink(linkUrl, idx)}
                      className="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-slate-700/50 rounded text-slate-400 hover:text-slate-200 shrink-0"
                      title="Copy URL"
                    >
                      {copiedIndex === idx ? (
                        <Check className="w-3.5 h-3.5 text-emerald-400" />
                      ) : (
                        <Copy className="w-3.5 h-3.5" />
                      )}
                    </button>
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}