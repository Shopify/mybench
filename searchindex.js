Search.setIndex({"docnames": ["detailed-design-doc", "eventrate-and-concurrency-control", "getting-started", "index", "introduction", "writing-a-benchmark"], "filenames": ["detailed-design-doc.rst", "eventrate-and-concurrency-control.rst", "getting-started.rst", "index.rst", "introduction.rst", "writing-a-benchmark.rst"], "titles": ["Detailed design documentation", "Event rate and concurrency control", "Getting started", "mybench documentation", "What is mybench?", "Tutorial: Writing and running a custom benchmark"], "terms": {"novemb": 0, "2022": 0, "databas": [0, 1, 2, 4, 5], "analysi": [0, 4, 5], "i": [0, 1, 2, 3, 5], "difficult": 0, "time": [0, 2, 4, 5], "consum": 0, "aspect": 0, "mainten": 0, "evolut": 0, "todai": 0, "": [0, 1, 2, 5], "cloud": 0, "base": [0, 1, 4, 5], "applic": 0, "One": [0, 5], "common": [0, 1], "method": [0, 5], "conduct": 0, "simul": 0, "load": [0, 2, 5], "match": [0, 5], "what": [0, 3, 5], "observ": [0, 5], "product": [0, 5], "ar": [0, 1, 4, 5], "typic": [0, 5], "against": [0, 1, 5], "server": [0, 4], "creat": [0, 3, 4], "environ": [0, 5], "similar": [0, 2, 4, 5], "specif": [0, 4, 5], "topologi": 0, "A": [0, 5], "set": [0, 1, 4, 5], "final": 0, "workload": [0, 1, 2, 3, 4], "sent": 0, "The": [0, 1, 2, 4, 5], "result": [0, 4, 5], "throughput": [0, 1, 4, 5], "latenc": [0, 4, 5], "metric": 0, "record": 0, "subsequ": 0, "analyz": [0, 4], "identifi": 0, "system": [0, 4, 5], "opportun": 0, "optim": 0, "To": [0, 1, 2, 5], "properli": [0, 5], "must": [0, 1, 2, 5], "realist": 0, "repres": [0, 5], "sinc": [0, 1, 5], "most": [0, 1, 5], "modern": 0, "web": [0, 4, 5], "somewhat": 0, "uniqu": [0, 5], "custom": [0, 3, 4], "model": [0, 3], "accur": [0, 4], "each": [0, 1, 4, 5], "mani": 0, "consist": [0, 5], "number": [0, 1, 4, 5], "microservic": 0, "its": [0, 1, 2, 4, 5], "own": [0, 1, 4, 5], "codebas": 0, "queri": [0, 5], "pattern": [0, 5], "gain": 0, "insight": 0, "whole": 0, "therefor": 0, "provid": [0, 5], "an": [0, 2, 5], "easi": 0, "us": [0, 1, 2, 4, 5], "ergonom": 0, "api": 0, "thi": [0, 1, 2, 4, 5], "term": 0, "pre": 0, "stage": 0, "1": [0, 1, 2, 5], "follow": [0, 1, 5], "execut": [0, 1, 5], "techniqu": 0, "discov": [0, 5], "drive": 0, "increas": 0, "until": 0, "satur": 0, "spike": 0, "precis": [0, 4], "which": [0, 1, 4, 5], "As": [0, 5], "amount": 0, "ha": [0, 4, 5], "do": [0, 4, 5], "also": [0, 4, 5], "can": [0, 1, 2, 4, 5], "neg": 0, "artifici": 0, "affect": 0, "avoid": [0, 5], "itself": [0, 5], "veri": [0, 4, 5], "effici": 0, "ideal": 0, "detect": 0, "compromis": 0, "desir": [0, 1, 5], "shorten": 0, "feedback": [0, 4], "loop": [0, 1, 4], "case": [0, 1, 5], "someth": [0, 2], "wrong": 0, "setup": [0, 5], "import": [0, 1, 5], "valu": [0, 1, 5], "engin": 0, "problem": [0, 1], "while": [0, 4, 5], "progress": 0, "abort": [0, 2], "run": [0, 1, 2, 3, 4], "necessari": [0, 5], "onc": [0, 2, 5], "complet": [0, 4], "larg": [0, 1], "would": 0, "thei": [0, 4, 5], "collect": [0, 4], "through": [0, 1, 2, 4], "should": [0, 1, 2, 4, 5], "visual": [0, 4, 5], "standard": [0, 4, 5], "manner": 0, "from": [0, 1, 4, 5], "differ": [0, 4, 5], "more": [0, 1, 2, 4, 5], "easili": [0, 4, 5], "compar": [0, 4, 5], "thu": [0, 5], "speed": 0, "up": [0, 1, 5], "interpret": 0, "sysbench": [0, 4], "sysbench01": 0, "benchbas": [0, 4], "benchbase01": 0, "linkbench": 0, "linkbench01": 0, "avail": 0, "none": [0, 1, 5], "fulfil": 0, "all": [0, 1, 3, 4], "outlin": 0, "abov": [0, 1, 5], "present": [0, 5], "here": [0, 1, 5], "our": [0, 4, 5], "attempt": [0, 5], "solv": 0, "singl": [0, 1, 4, 5], "softwar": 0, "packag": [0, 5], "we": [0, 5], "name": 0, "author": 0, "borrow": 0, "terminologi": 0, "solid": 0, "fluid": 0, "dynam": 0, "In": [0, 1, 4, 5], "those": [0, 5], "field": [0, 5], "where": [0, 4, 5], "physic": [0, 5], "being": [0, 4, 5], "when": [0, 1, 5], "calcul": [0, 1], "made": [0, 1, 4], "industri": 0, "includ": [0, 4], "help": [0, 4], "obtain": [0, 5], "faster": [0, 4], "http": [0, 2, 4, 5], "github": [0, 2, 5], "com": [0, 2, 5], "akopytov": 0, "cmu": 0, "db": 0, "facebookarch": 0, "were": 0, "develop": [0, 4, 5], "dure": [0, 4, 5], "get": [0, 3, 5], "out": [0, 5], "wai": [0, 4, 5], "possibl": [0, 4], "conveni": 0, "encourag": 0, "code": [0, 5], "impact": 0, "multipl": [0, 1, 4, 5], "mix": [0, 4], "better": [0, 4], "emul": 0, "traffic": [0, 5], "group": 0, "exist": [0, 4, 5], "k": 0, "event": [0, 2, 3, 4, 5], "statist": [0, 5], "seri": [0, 4, 5], "format": 0, "plot": [0, 4, 5], "golang": [0, 1, 5], "librari": [0, 4, 5], "enabl": [0, 4], "minim": 0, "effort": 0, "defin": [0, 2, 3, 4], "top": [0, 5], "level": 0, "kept": 0, "small": 0, "much": [0, 5], "complex": [0, 5], "within": [0, 5], "figur": [0, 5], "depict": 0, "show": [0, 2, 5], "3": 0, "two": [0, 5], "benchmarkwork": 0, "At": 0, "least": [0, 1, 5], "one": [0, 1, 4, 5], "callback": 0, "contain": [0, 5], "For": [0, 1, 5], "exampl": [0, 1, 2, 5], "could": [0, 5], "sequenc": [0, 5], "request": [0, 5], "particular": [0, 5], "endpoint": [0, 5], "anoth": 0, "three": [0, 5], "togeth": [0, 3], "configur": [0, 5], "parallel": [0, 4], "overal": [0, 5], "split": [0, 2, 5], "worker": [0, 1, 5], "scenario": 0, "connect": [0, 1, 2, 4, 5], "send": 0, "simultan": 0, "goroutin": [0, 1, 4, 5], "equal": 0, "divid": 0, "between": [0, 1, 2, 5], "emb": [0, 5], "object": [0, 5], "see": [0, 1, 2, 5], "onlinehistogram": 0, "store": [0, 5], "held": 0, "insid": 0, "period": [0, 5], "read": [0, 5], "datalogg": 0, "aggreg": [0, 4], "across": [0, 4], "per": [0, 1, 5], "These": [0, 5], "written": [0, 4], "sqlite": [0, 2, 4, 5], "need": [0, 1, 2, 5], "concern": [0, 5], "struct": [0, 5], "workloadinterfac": [0, 3], "definit": 0, "config": [0, 2, 5], "return": [0, 5], "newcontextdata": [0, 5], "construct": [0, 4], "thread": [0, 5], "local": [0, 5], "structur": [0, 3], "state": [0, 5], "ti": 0, "benchmarkapp": 0, "ad": 0, "handl": [0, 5], "other": [0, 3, 5], "administr": 0, "duti": 0, "pars": [0, 5], "command": [0, 5], "line": [0, 5], "flag": [0, 5], "approach": [0, 4], "overload": 0, "plateau": [0, 5], "declin": 0, "type": [0, 5], "naiv": 0, "sleep": 0, "simpl": [0, 5], "durat": [0, 2, 4, 5], "howev": [0, 4], "maintain": [0, 1, 4], "beyond": 0, "500": [0, 1], "1000": [0, 1, 2], "hz": [0, 1], "due": [0, 1], "schedul": [0, 1], "incur": 0, "linux": [0, 1], "without": 0, "real": [0, 2, 4], "patch": 0, "appli": 0, "addition": [0, 5], "go": [0, 1, 2, 4, 5], "introduc": 0, "addit": [0, 4], "order": [0, 5], "about": [0, 2, 5], "m": 0, "test": [0, 2, 4], "further": [0, 4, 5], "difficulti": 0, "instead": [0, 5], "call": [0, 1, 2, 5], "iter": [0, 1], "nest": 0, "inner": 0, "outer": [0, 1], "constant": [0, 1], "rel": [0, 1, 5], "low": 0, "frequenc": [0, 5], "50": [0, 1, 5], "default": [0, 1, 2, 4, 5], "arriv": 0, "sampl": [0, 4, 5], "either": [0, 4], "uniform": [0, 5], "poisson": 0, "distribut": [0, 2, 5], "2a": 0, "after": [0, 4, 5], "next": 0, "wake": 0, "accord": [0, 1, 5], "repeat": 0, "same": [0, 4, 5], "2": [0, 1, 5], "scheme": 0, "normal": [0, 5], "oper": [0, 1], "circumst": 0, "b": [0, 5], "too": 0, "slow": [0, 1, 5], "keep": 0, "box": 0, "long": [0, 4], "overrun": 0, "switch": 0, "mode": 0, "2b": 0, "onli": [0, 3, 4, 5], "success": 0, "remov": [0, 5], "maxim": [0, 1], "effect": 0, "track": 0, "have": [0, 1, 4, 5], "start": [0, 3, 5], "actual": [0, 5], "If": [0, 1, 5], "lower": 0, "than": [0, 1, 4, 5], "ani": [0, 5], "point": [0, 5], "catch": 0, "expect": 0, "back": [0, 4, 5], "batch": [0, 1, 5], "allow": [0, 4, 5], "averag": [0, 5], "over": [0, 4], "momentarili": 0, "threshold": 0, "longer": [0, 5], "sustain": 0, "slightli": 0, "current": [0, 5], "mai": [0, 1, 5], "updat": [0, 5], "descript": 0, "correct": 0, "todo": [0, 2], "hdr": 0, "histogram": [0, 5], "hdrhist01": 0, "memori": [0, 5], "cpu": [0, 5], "overhead": 0, "instanc": [0, 5], "embed": [0, 5], "capabl": [0, 4], "everi": 0, "second": [0, 1, 5], "cours": 0, "continu": 0, "access": [0, 5], "modifi": 0, "respect": [0, 5], "race": [0, 5], "main": [0, 3, 4], "hundr": [0, 5], "short": 0, "enough": 0, "bia": 0, "satisfi": [0, 5], "underli": 0, "write": [0, 3, 4], "activ": 0, "slot": 0, "3a": 0, "logger": 0, "take": [0, 5], "snapshot": 0, "swap": 0, "inact": 0, "3b": 0, "mutex": [0, 5], "guard": 0, "both": [0, 1, 4, 5], "fast": 0, "simpli": 0, "index": [0, 3, 5], "immedi": 0, "unblock": 0, "now": [0, 5], "newli": 0, "3c": 0, "occur": [0, 5], "resid": 0, "reset": 0, "zero": 0, "reus": 0, "idl": 0, "c": [0, 2], "disk": [0, 5], "occupi": 0, "tabl": [0, 2], "meta": 0, "inform": 0, "stop": 0, "choic": 0, "file": [0, 2, 4, 5], "simplifi": [0, 5], "transport": 0, "storag": [0, 5], "hdrhistogram": 0, "org": [0, 5], "elaps": 0, "sum": [0, 5], "count": 0, "fraction": 0, "signific": 0, "numer": 0, "denomin": 0, "passag": 0, "built": [0, 4, 5], "alreadi": [0, 5], "claus": [0, 4, 5], "select": [0, 4, 5], "statement": [0, 4, 5], "primit": 0, "intens": 0, "doe": [0, 5], "fulli": [0, 1], "version": [0, 5], "algorithm": [0, 5], "done": [0, 2, 4, 5], "best": 0, "ensur": [0, 5], "footprint": 0, "random": [0, 4, 5], "rand": [0, 5], "global": [0, 5], "sourc": 0, "protect": 0, "concurr": [0, 3, 5], "bottleneck": 0, "elimin": 0, "displai": 0, "five": 0, "gather": [0, 4], "ring": 0, "vegalit": 0, "vega01": 0, "issu": [0, 5], "quickli": 0, "4": [0, 1], "screenshot": 0, "vega": 0, "io": 0, "lite": 0, "suit": 0, "them": [0, 5], "5": [0, 5], "comparison": [0, 3], "d": [0, 5], "achiev": [1, 5], "high": [1, 3, 4, 5], "total": [1, 5], "mybench": [1, 2, 5], "util": 1, "benchmark": [1, 2, 3, 4], "via": [1, 4, 5], "thousand": 1, "section": 1, "detail": [1, 3, 5], "design": [1, 3, 5], "document": [1, 5], "paramet": [1, 5], "workermaxr": 1, "max": [1, 5], "so": [1, 4, 5], "automat": [1, 2], "argument": [1, 5], "determin": [1, 5], "100": [1, 5], "few": 1, "how": [1, 2, 5], "tune": 1, "formula": 1, "ceil": 1, "spread": 1, "workloadconfig": [1, 5], "workloadscal": [1, 5], "greater": 1, "perform": [1, 3, 4, 5], "round": 1, "illustr": 1, "purpos": 1, "some": [1, 5], "auto": 1, "10000": [1, 2], "35000": 1, "350": 1, "35001": 1, "71": 1, "adjust": 1, "suffici": 1, "target": 1, "situat": [1, 5], "fix": [1, 2], "wish": [1, 5], "ignor": [1, 5], "By": [1, 2, 4, 5], "25": [1, 5], "variabl": [1, 5], "behavior": [1, 5], "looper": 1, "errat": 1, "independ": 1, "although": [1, 4], "guarante": 1, "root": 1, "gener": [1, 3, 4], "exce": 1, "200": [1, 5], "unlik": 1, "abl": [1, 4], "divis": 1, "steadili": 1, "experi": 1, "oscil": 1, "around": [1, 4, 5], "discret": [1, 4], "error": [1, 5], "guid": 2, "walk": 2, "you": [2, 5], "instal": [2, 5], "implement": [2, 3, 4], "download": 2, "compil": 2, "examplebench": 2, "git": 2, "clone": 2, "shopifi": [2, 5], "cd": [2, 5], "make": [2, 5], "folder": [2, 5], "build": [2, 4, 5], "first": [2, 5], "seed": [2, 5], "host": [2, 5], "mysql": [2, 4, 5], "user": [2, 5], "usernam": 2, "pass": [2, 5], "password": 2, "replac": [2, 5], "ip": 2, "address": 2, "million": [2, 5], "row": [2, 5], "data": [2, 3, 4], "example_t": 2, "bench": [2, 5], "eventr": [2, 3, 5], "rate": [2, 3, 4, 5], "evenli": 2, "variou": 2, "option": 2, "overrid": 2, "specifi": [2, 3, 5], "10x": 2, "among": 2, "localhost": [2, 5], "8005": [2, 5], "monitor": [2, 4, 5], "ui": [2, 4], "indefinit": [2, 5], "press": 2, "ctrl": 2, "save": [2, 5], "bit": 2, "post": [2, 3, 4], "process": [2, 3, 4], "script": [2, 4, 5], "framework": [3, 4], "rapid": 3, "prototyp": 3, "tool": [3, 5], "tutori": 3, "project": 3, "benchmarkinterfac": 3, "put": 3, "trace": 3, "review": 3, "control": [3, 4, 5], "advanc": 3, "chang": 3, "outerloopr": 3, "modul": [3, 5], "search": 3, "page": 3, "massiv": 4, "primarili": 4, "support": 4, "even": [4, 5], "non": 4, "featur": 4, "arbitrari": [4, 5], "ratio": 4, "individu": 4, "expos": 4, "interfac": [4, 5], "valid": [4, 5], "log": [4, 5], "share": 4, "interoper": 4, "well": [4, 5], "easier": [4, 5], "perfect": 4, "moment": 4, "lua": 4, "It": [4, 5], "histori": 4, "commun": 4, "de": 4, "facto": 4, "date": 4, "2004": 4, "significantli": 4, "intern": [4, 5], "multi": 4, "manual": [4, 5], "dbm": 4, "java": 4, "wherea": 4, "focus": 4, "heavi": 4, "jdbc": 4, "principl": 4, "work": 4, "evolv": 4, "mixtur": 4, "cannot": 4, "nativ": 4, "requir": 5, "note": 5, "architectur": 5, "doc": 5, "function": 5, "still": 5, "fundament": 5, "suitabl": 5, "care": 5, "examin": 5, "constraint": 5, "explain": 5, "creation": 5, "outsid": 5, "scope": 5, "aim": 5, "teach": 5, "microblog": 5, "servic": 5, "chirp": 5, "befor": 5, "concept": 5, "column": 5, "id": 5, "bigint": 5, "20": 5, "content": 5, "varchar": 5, "140": 5, "created_at": 5, "datetim": 5, "indic": 5, "primari": 5, "kei": 5, "There": 5, "latest": 5, "75": 5, "BY": 5, "desc": 5, "limit": 5, "third": 5, "new": 5, "insert": 5, "INTO": 5, "null": 5, "length": 5, "predefin": 5, "discuss": 5, "later": 5, "scale": 5, "platform": 5, "assum": 5, "10": 5, "found": 5, "step": 5, "involv": 5, "mod": 5, "mkdir": 5, "tutorialbench": 5, "init": 5, "let": 5, "four": 5, "program": 5, "correspond": 5, "table_chirp": 5, "workload_read_latest_chirp": 5, "workload_read_single_chirp": 5, "workload_insert_chirp": 5, "func": 5, "newtablechirp": 5, "idgen": 5, "datagener": 5, "initializet": 5, "NOT": 5, "auto_incr": 5, "newhistogramlengthstringgener": 5, "float64": 5, "0": 5, "40": 5, "60": 5, "80": 5, "120": 5, "15": 5, "newnowgener": 5, "string": 5, "primarykei": 5, "phase": 5, "samplefromexist": 5, "becaus": 5, "behaviour": 5, "like": 5, "arrai": 5, "binsendpoint": 5, "bin": 5, "suppos": 5, "integ": 5, "119": 5, "charact": 5, "filter": 5, "don": 5, "t": 5, "worri": 5, "attribut": 5, "safe": 5, "percentag": 5, "alloc": 5, "databaseconfig": 5, "context": 5, "fmt": 5, "nocontextdata": 5, "w": 5, "ctx": 5, "workercontext": 5, "sprintf": 5, "_": 5, "err": 5, "conn": 5, "newreadlatestchirp": 5, "abstractworkload": 5, "newworkload": 5, "hold": 5, "thin": 5, "wrapper": 5, "client": 5, "reason": 5, "safeti": 5, "report": 5, "express": 5, "parameter": 5, "upon": 5, "describ": 5, "randomli": 5, "exercis": 5, "alwai": 5, "prepar": 5, "associ": 5, "readsinglechirpcontext": 5, "stmt": 5, "var": 5, "newreadsinglechirp": 5, "try": 5, "reentrant": 5, "interest": 5, "export": 5, "down": 5, "lastli": 5, "last": 5, "createdat": 5, "newinsertchirp": 5, "05": 5, "rememb": 5, "complic": 5, "list": 5, "ran": 5, "entir": 5, "chirpbench": 5, "benchmarkconfig": 5, "initialnumrow": 5, "int64": 5, "newautoincrementgener": 5, "reloaddata": 5, "ratecontrolconfig": 5, "nil": 5, "newautoincrementgeneratorfromdatabas": 5, "everyth": 5, "except": 5, "popul": 5, "given": 5, "usual": 5, "leverag": 5, "drop": 5, "recreat": 5, "shown": 5, "autoincrementgener": 5, "min": 5, "minimum": 5, "atom": 5, "increment": 5, "size": 5, "chosen": 5, "good": 5, "constructor": 5, "place": 5, "slice": 5, "again": 5, "maximum": 5, "With": 5, "newbenchmarkconfig": 5, "int64var": 5, "numrow": 5, "10_000_000": 5, "panic": 5, "fill": 5, "loader": 5, "o": 5, "ll": 5, "your": 5, "57": 5, "sy": 5, "admin_rw": 5, "hunter2": 5, "info": 5, "0000": 5, "reload": 5, "batchsiz": 5, "16": 5, "totalrow": 5, "10000000": 5, "pct": 5, "rowsinsert": 5, "100400": 5, "0056": 5, "99": 5, "19": 5, "9919400": 5, "0057": 5, "8": 5, "750": 5, "live": 5, "shell": 5, "minut": 5, "goal": 5, "sh": 5, "bash": 5, "2m": 5, "3500": 5, "4000": 5, "4500": 5, "5000": 5, "5500": 5, "6000": 5, "fresh": 5, "end": 5, "rm": 5, "f": 5, "append": 5, "pertain": 5, "copi": 5, "repositori": 5, "Then": 5, "jupyt": 5, "notebook": 5, "ipynb": 5, "depend": 5, "yml": 5, "conda": 5, "cell": 5, "look": 5, "qp": 5, "bottom": 5, "percentil": 5, "6": 5, "bar": 5, "x": 5, "sigma": 5, "deviat": 5, "color": 5, "annot": 5, "right": 5, "reach": 5, "shift": 5, "higher": 5, "extra": 5, "skew": 5, "runtim": 5, "learn": 5, "basic": 5, "sql": 5, "vari": 5, "ascertain": 5, "python": 5, "histogramlengthstringgener": 5}, "objects": {}, "objtypes": {}, "objnames": {}, "titleterms": {"detail": 0, "design": 0, "document": [0, 3], "mybench": [0, 3, 4], "high": 0, "perform": 0, "framework": 0, "rapid": 0, "prototyp": 0, "benchmark": [0, 5], "introduct": 0, "requir": 0, "intern": 0, "architectur": 0, "rate": [0, 1], "control": [0, 1], "via": 0, "tempor": 0, "discret": 0, "looper": 0, "data": [0, 5], "log": 0, "doubl": 0, "buffer": 0, "gener": [0, 5], "live": 0, "monitor": 0, "user": 0, "interfac": 0, "post": [0, 5], "process": [0, 5], "tool": [0, 4], "experiment": 0, "evalu": 0, "discuss": 0, "stabil": 0, "resourc": 0, "util": 0, "eas": 0, "implement": [0, 5], "new": 0, "limit": 0, "futur": 0, "work": 0, "conclus": 0, "event": 1, "concurr": 1, "specifi": 1, "onli": 1, "eventr": 1, "advanc": 1, "chang": 1, "outerloopr": 1, "get": 2, "start": 2, "content": 3, "indic": 3, "tabl": [3, 5], "what": 4, "i": 4, "comparison": 4, "other": 4, "tutori": 5, "write": 5, "run": 5, "custom": 5, "model": 5, "workload": 5, "definit": 5, "initi": 5, "creat": 5, "project": 5, "structur": 5, "defin": 5, "workloadinterfac": 5, "readlatestchirp": 5, "readsinglechirp": 5, "insertchirp": 5, "benchmarkinterfac": 5, "name": 5, "runload": 5, "put": 5, "all": 5, "togeth": 5, "main": 5, "trace": 5, "review": 5}, "envversion": {"sphinx.domains.c": 2, "sphinx.domains.changeset": 1, "sphinx.domains.citation": 1, "sphinx.domains.cpp": 8, "sphinx.domains.index": 1, "sphinx.domains.javascript": 2, "sphinx.domains.math": 2, "sphinx.domains.python": 3, "sphinx.domains.rst": 2, "sphinx.domains.std": 2, "sphinx": 57}, "alltitles": {"Detailed design documentation": [[0, "detailed-design-documentation"]], "mybench: a high-performance framework for rapid prototyping of benchmarks": [[0, "mybench-a-high-performance-framework-for-rapid-prototyping-of-benchmarks"]], "Introduction": [[0, "introduction"]], "Requirements": [[0, "requirements"]], "mybench design": [[0, "mybench-design"]], "Internal architecture": [[0, "internal-architecture"]], "Rate control via \u201ctemporal-discretization\u201d looper": [[0, "rate-control-via-temporal-discretization-looper"]], "High-performance data logging via double buffering": [[0, "high-performance-data-logging-via-double-buffering"]], "Data generation": [[0, "data-generation"]], "Live monitoring user interface": [[0, "live-monitoring-user-interface"]], "Post-processing tools": [[0, "post-processing-tools"]], "Experimental evaluations and discussions": [[0, "experimental-evaluations-and-discussions"]], "Rate control stability": [[0, "rate-control-stability"]], "mybench resource utilization": [[0, "mybench-resource-utilization"]], "Ease of implementation of new benchmarks": [[0, "ease-of-implementation-of-new-benchmarks"]], "Limitations and future work": [[0, "limitations-and-future-work"]], "Conclusion": [[0, "conclusion"]], "Event rate and concurrency control": [[1, "event-rate-and-concurrency-control"]], "Specifying only -eventrate": [[1, "specifying-only-eventrate"]], "Specifying -eventrate and -concurrency": [[1, "specifying-eventrate-and-concurrency"]], "Advanced: changing -outerlooprate": [[1, "advanced-changing-outerlooprate"]], "Getting started": [[2, "getting-started"]], "mybench documentation": [[3, "mybench-documentation"]], "Contents:": [[3, null]], "Indices and tables": [[3, "indices-and-tables"]], "What is mybench?": [[4, "what-is-mybench"]], "Comparisons with other tools": [[4, "comparisons-with-other-tools"]], "Tutorial: Writing and running a custom benchmark": [[5, "tutorial-writing-and-running-a-custom-benchmark"]], "Modeling the workload": [[5, "modeling-the-workload"]], "Table definition": [[5, "table-definition"]], "Workload definition": [[5, "workload-definition"]], "Initial data": [[5, "initial-data"]], "Creating a project structure": [[5, "creating-a-project-structure"]], "Defining the table and data generators": [[5, "defining-the-table-and-data-generators"]], "Implementing WorkloadInterface": [[5, "implementing-workloadinterface"]], "ReadLatestChirps": [[5, "readlatestchirps"]], "ReadSingleChirp": [[5, "readsinglechirp"]], "InsertChirp": [[5, "insertchirp"]], "Implementing BenchmarkInterface": [[5, "implementing-benchmarkinterface"]], "Name()": [[5, "name"]], "RunLoader()": [[5, "runloader"]], "Workloads()": [[5, "workloads"]], "Putting it all together in main()": [[5, "putting-it-all-together-in-main"]], "Running the benchmark": [[5, "running-the-benchmark"]], "Post processing the data": [[5, "post-processing-the-data"]], "Tracing the benchmark": [[5, "tracing-the-benchmark"]], "Review": [[5, "review"]]}, "indexentries": {}})