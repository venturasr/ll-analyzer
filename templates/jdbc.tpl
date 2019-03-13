<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <title>Report: jdbc log</title>
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet" />
    <link href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.0.0-beta/css/bootstrap.min.css" rel="stylesheet" />
    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.6/css/bootstrap.min.css" integrity="sha384-rwoIResjU2yc3z8GV/NPeZWAv56rSmLldC3R/AZzGRnGxQQKnKkoFVhFQhNUwEyJ" crossorigin="anonymous">
    <style type="text/css" media="screen">
        body {
            font: normal 14px/1.2 Tahoma, Geneva, sans-serif;
            color: #000000;
        }

        div.container {
            padding: 20px;
            padding-top: 20px;
        }

        div.col {
            padding-top: .50rem;
            padding-bottom: .50rem;
            background: #ffffff;
            border: 1px solid rgba(2, 2, 2, 0.15);
            -webkit-box-sizing: content-box;
            -moz-box-sizing: content-box;
            box-sizing: content-box;
            webkit-box-shadow: 1px 1px 1px 0 rgba(0, 0, 0, 0.3);
            box-shadow: 2px 2px 2px 0 rgba(0, 0, 0, 0.3);
            overflow: hidden;
            width: auto;
            word-wrap: break-word;
        }

        div.col-6.col-md-3 {
            padding-top: .75rem;
            padding-bottom: .75rem;
            border: 1px solid rgba(2, 2, 2, 0.15);
            webkit-box-shadow: 1px 1px 1px 0 rgba(0, 0, 0, 0.3);
            box-shadow: 2px 2px 2px 0 rgba(0, 0, 0, 0.3);
            overflow: hidden;
            width: auto;
            word-wrap: break-word;
        }

        div.col-12.col-sm-6.col-md-9 {
            padding-top: .75rem;
            padding-bottom: .75rem;
            border: 1px solid rgba(2, 2, 2, 0.15);
            webkit-box-shadow: 1px 1px 1px 0 rgba(0, 0, 0, 0.3);
            box-shadow: 2px 2px 2px 0 rgba(0, 0, 0, 0.3);
            overflow: hidden;
            width: auto;
            word-wrap: break-word;
        }

        div#statement {
            background: #0388bc;
            font: normal bold 16px/1.2 Tahoma, Geneva, sans-serif;
            color: #ffffff;
            padding-top: .75rem;
            padding-bottom: .75rem;
        }

        div.col.Accordion {
            padding-top: 0rem;
            padding-bottom: 0rem;
            padding-right: 0px;
            padding-left: 0px;
        }


        div.CellHead {
            background: #ffffff;
            font: normal bold 12px/1.2 Tahoma, Geneva, sans-serif;
            color: #000000;
            width: auto;
            text-align: left;
        }

        div.Accordion a {
            background: #ffffff;
            font: normal bold 14px/1.2 Tahoma, Geneva, sans-serif;
            color: #000000;
            width: auto;
            text-align: center;
        }

        div.card-header {
            border: 0px solid rgba(0, 0, 0, .125);
            border-radius: 0rem;
        }


        div.card {
            background-color: #fff;
            border: 0px solid rgba(0, 0, 0, .125);
            border-radius: 0rem;
        }


        [data-toggle="collapse"]:after {
            display: inline-block;
            display: inline-block;
            font: normal normal normal 14px/1 FontAwesome;
            font-size: inherit;
            text-rendering: auto;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
            content: "\f054";
            transform: rotate(90deg);
            transition: all linear 0.25s;
            float: right;
        }

        [data-toggle="collapse"].collapsed:after {
            transform: rotate(0deg);
        }
    </style>
</head>

<body>

<div class="container">
    <div><h1>Report: {{.ReportType}} log</h1></div>
    <div><h3>Number of logs analyzed {{.NumberOfFiles}}<h3></div>
</div>

{{with .Issues}}
{{range .}}
<!-- Statement -->
<div class="container">
    <div class="row">
        <div id="statement" class="col-6 col-md-3">
            Statement
        </div>
        <div class="col-12 col-sm-6 col-md-9">
        {{.Statement}}
        </div>
    </div>
    </br>
    <!-- Summary -->
    <div class="row">
        <div class="col CellHead">
            Occurrences: {{.Occurrences}}
        </div>
        <div class="col CellHead">
            Higher Time:  {{.HigherTimeMillis}} ms
        </div>
        <div class="col CellHead">
            Average Time: {{.AverageTimeMillis}} ms
        </div>
        <div class="col CellHead">
            Total Time: {{.TotalTime}} ms
        </div>

        <div class="col CellHead">
            Files: {{with .IssueFileNames}}{{range .}}{{.}}{{end}}{{end}}
        </div>

    </div>

    <!-- SQL -->
    <div class="row">
    {{with .Sqls}}
    {{range $index, $results := .}}
    {{$collapse := uniqueCollapse $index}}
        <div class="col Accordion">
            <div id="accordion1" role="tablist">
                <div class="card">
                    <div class="card-header" role="tab" id="headingOne">
                        <h5 class="mb-0">
                            <a class="collapsed" data-toggle="collapse" href="#collapsed{{$collapse}}"  aria-expanded="false" aria-controls="collapsed{{$collapse}}">
                                SQL
                            </a>
                        </h5>
                    </div>
                    <div id="collapsed{{$collapse}}" class="collapse" role="tabpanel" aria-labelledby="headingOne">
                        <div class="card-body">
                        {{.Sql}}
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="col Accordion">
            <div id="accordion2" role="tablist">
                <div class="card">
                    <div class="card-header" role="tab" id="headingTwo">
                        <h5 class="mb-0">
                            <a class="collapsed" data-toggle="collapse" href="#collapsed{{$collapse}}1"  data-target="#collapsed{{$collapse}}1" aria-expanded="false" aria-controls="collapsed{{$collapse}}1">
                                TRACE
                            </a>
                        </h5>
                    </div>
                    <div id="collapsed{{$collapse}}1" class="collapse" role="tabpanel" aria-labelledby="headingTwo">
                        <div class="card-body">
                        {{with .Trace}}
                            {{range .}}
                            <p>{{.}}</p>
                            {{end}}
                        {{end}}
                        </div>
                    </div>
                </div>
            </div>
        </div>
            {{end}}
            {{end}}
    </div>
</div>

{{end}}
{{end}}

<!-- jQuery first, then Tether, then Bootstrap JS. -->
<script src="https://code.jquery.com/jquery-3.1.1.slim.min.js" integrity="sha384-A7FZj7v+d/sdmMqp/nOQwliLvUsJfDHW+k9Omg/a/EheAdgtzNs3hpfag6Ed950n" crossorigin="anonymous"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.4.0/js/tether.min.js" integrity="sha384-DztdAPBWPRXSA/3eYEEUWrWCy7G5KFbe8fFjk5JAIxUYHKkDx6Qin1DkWx51bBrb" crossorigin="anonymous"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.6/js/bootstrap.min.js" integrity="sha384-vBWWzlZJ8ea9aCX4pEW3rVHjgjt7zpkNpZk+02D9phzyeVkE+jo0ieGizqPLForn" crossorigin="anonymous"></script>
</body>

</html>