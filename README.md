Serverless service for ZEIT. Written in [Golang](https://golang.org/).

## Deploy your own

You'll want to fork this repository and deploy your own image to primitive.

1. Click the fork button at the top right of GitHub
2. Clone the repo to your local machine with `git clone URL_OF_FORKED_REPO_HERE`
4. Deploy by running `now` from the CLI (if you don't already have it, run `npm install -g now`)

Alternatively, you can do a one-click to deploy with the button below.

[![Deploy to now](https://deploy.now.sh/static/button.svg)](https://zeit.co/new/project?template=xuthus5/zeit-serverless)


  多个源文件可同属于一个包，只要声明时package指定的包名一样；
  一个包对应生成一个*.a文件，生成的文件名并不是包名+.a组成，应该是目录名+.a组成
  go install ××× 这里对应的并不是包名，而是路径名！！
  import ××× 这里使用的也不是包名，也是路径名
  ×××××.SayHello() 这里使用的才是包名！
  指定×××路径名就代表了此目录下唯一的包，编译器连接器默认就会去生成或者使用它，而不需要我们手动指明！
  一个目录下就只能有一个包存在
  对于调用有源码的第三方包，连接器在连接时，其实使用的并不是我们工作目录下的.a文件，而是以该最新源码编译出的临时文件夹中的.a文件
  对于调用没有源码的第三方包，上面的临时编译不可能成功，那么临时目录下就不可能有.a文件，所以最后链接时就只能链接到工作目录下的.a文件
  对于标准库，即便是修改了源代码，只要不重新编译Go源码，那么链接时使用的就还是已经编译好的*.a文件
  包导入有三种模式：正常模式、别名模式、简便模式
