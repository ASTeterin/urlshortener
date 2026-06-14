# internal/service

Пакет `service` содержит бизнес-логику и играет ключевую роль в реализации функциональности приложения.

В нём описаны правила, процессы и операции, которые определяют поведение приложения.

Принципы организации:
- Сервисы должны быть независимы от деталей транспорта (HTTP, gRPC и т.д.).
- Взаимодействие с базой данных происходит через интерфейсы репозиториев.
- Каждый сервис должен иметь четко определенную область ответственности.

Результаты оптимизации
File: service.test
Build ID: cd5681fbbeb0428c1874c6ff4c6f4726e2d3de9b
Type: alloc_space
Time: 2026-06-11 08:36:40 MSK
Showing nodes accounting for 1.10GB, 25.59% of 4.30GB total
Dropped 53 nodes (cum <= 0.02GB)
flat  flat%   sum%        cum   cum%
-1.48GB 34.45% 34.45%    -1.48GB 34.45%  math/rand.newSource (inline)
1.18GB 27.55%  6.90%     1.18GB 27.55%  github.com/ASTeterin/urlshortener/internal/service.(*mockRepo).Store
0.71GB 16.45%  9.55%     0.65GB 15.18%  github.com/ASTeterin/urlshortener/internal/service.(*mockRepo).StoreAll
-0.42GB  9.74%  0.19%    -0.04GB  0.91%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).ListUserURLs
0.41GB  9.61%  9.42%     0.41GB  9.61%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).GetOriginalURL
0.38GB  8.82% 18.25%     0.38GB  8.82%  github.com/ASTeterin/urlshortener/internal/service.(*mockRepo).ListByUserID
0.29GB  6.84% 25.09%     0.25GB  5.77%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).ListShortURL
0.11GB  2.61% 27.70%    -0.70GB 16.25%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).generateModels
-0.07GB  1.70% 26.00%    -0.13GB  3.11%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).BatchRemove
-0.06GB  1.35% 24.65%    -0.06GB  1.35%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).splitIntoBatches (inline)
0.04GB  0.94% 25.59%    -1.44GB 33.50%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).generateShortURL
0     0% 25.59%     0.61GB 14.17%  github.com/ASTeterin/urlshortener/internal/service.(*shortenerService).GetShortURL
0     0% 25.59%    -0.02GB  0.45%  github.com/ASTeterin/urlshortener/internal/service.BenchmarkBatchRemove
0     0% 25.59%    -0.13GB  3.11%  github.com/ASTeterin/urlshortener/internal/service.BenchmarkBatchRemove.func1
0     0% 25.59%     0.41GB  9.61%  github.com/ASTeterin/urlshortener/internal/service.BenchmarkGetOriginalURL
0     0% 25.59%     0.65GB 15.13%  github.com/ASTeterin/urlshortener/internal/service.BenchmarkGetShortURL
0     0% 25.59%     0.25GB  5.77%  github.com/ASTeterin/urlshortener/internal/service.BenchmarkListShortURL
0     0% 25.59%    -0.06GB  1.42%  github.com/ASTeterin/urlshortener/internal/service.BenchmarkListUserURLs
0     0% 25.59%    -1.48GB 34.45%  math/rand.NewSource (inline)
0     0% 25.59%    -0.13GB  3.11%  testing.(*B).RunParallel.func1
0     0% 25.59%     1.25GB 29.07%  testing.(*B).launch
0     0% 25.59%    -0.02GB  0.41%  testing.(*B).run1.func1
0     0% 25.59%     1.23GB 28.67%  testing.(*B).runN
