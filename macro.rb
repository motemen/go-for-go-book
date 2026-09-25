require 'asciidoctor'
require 'asciidoctor/extensions'
require 'pathname'
require 'json'
require 'digest'
require 'uri'
require 'net/https'

CACHE_DIR    = Pathname.new('./.cache')
EXAMPLES_DIR = Pathname.new('./examples')

# bundler が vendor/bundle に gem を入れると、go コマンドがそれを Go の vendor
# ディレクトリと誤認して失敗する。go run はモジュールモードの明示で回避できるが、
# go doc は回避できないので、gem は vendor/bundle 以外に入れること。
ENV['GOFLAGS'] ||= '-mod=mod'

CONFIG     = JSON.parse(File.read('config.json'))
GO_VERSION = CONFIG['Versions']['Go']

# コマンドを実行して出力を返す。成功した結果だけを .cache/ にキャッシュする。
# 戻り値は [出力, 成功したか]。
def run_cached(kind, command, file = nil)
  cache_file = Pathname.new(CACHE_DIR+"#{kind}-#{GO_VERSION}"+(file || command.gsub(/[^\w.]/, '_')))
  cache_mtime = cache_file.mtime rescue Time.new(0)
  if File.exist?(cache_file) && !(file && cache_mtime < file.mtime)
    return [cache_file.read(encoding: Encoding::UTF_8), true]
  end

  STDERR.puts "macro: #{command}"
  c = %x(#{command}).force_encoding(Encoding::UTF_8)
  ok = $?.success?
  if ok
    cache_file.parent.mkpath
    cache_file.write(c)
  end
  [c, ok]
end

# ref: http://asciidoctor.org/docs/user-manual/#block-macro-processor-example

# Expands to an example code under ./examples or its output
# TODO: place a link to Go playground
#
#   goexample::parsefile[]
#   goexample::parsefile[output]
#
# Runs examples/parseexpr/parseexpr.go
class GoExampleMacro < Asciidoctor::Extensions::BlockMacroProcessor
  use_dsl

  named :goexample

  def process(parent, target, attrs)
    filename = attrs['file'] || "#{target}.go"
    file = EXAMPLES_DIR + "#{target}/#{filename}"
    style = attrs.delete(1)

    if style === 'output'
      content, ok = run_cached('go-run', "go run #{file} 2>&1", file)
      Asciidoctor::LoggerManager.logger.warn "goexample::#{target}[output] failed: #{content.lines.first&.chomp}" unless ok
      create_listing_block(
        parent,
        content,
        attrs
      )
    else
      source = IO.read(file)
      digest = Digest::SHA1.hexdigest(source)

      playground_keys_file = EXAMPLES_DIR + 'playground-keys.json'
      playground_keys = JSON.parse(File.read(playground_keys_file))

      playground_key = playground_keys[digest]
      unless playground_key
        # quick check if the example contains non-standard package or not
        if /\./ === %x(go list -f {{.Imports}} #{file}) # we know that file starts with ./
          # nop
        else
          STDERR.print "macro: sharing #{file} to playground ... "

          begin
            uri = URI('https://play.golang.org/share')
            http = Net::HTTP.new(uri.host, uri.port)
            http.use_ssl = true

            req = Net::HTTP::Post.new uri
            req.body = source
            req['Content-Type'] = 'text/plain'

            resp = http.request req

            if 200 <= resp.code.to_i && resp.code.to_i < 300
              playground_keys[digest] = playground_key = resp.body.chomp
              File.open(playground_keys_file, 'w') do |f|
                f.puts playground_keys.to_json
              end

              STDERR.puts "succeeded: #{playground_key}"
            else
              raise "got [#{resp.code} #{resp.message}]"
            end
          rescue => e
            STDERR.puts "failed: #{e}"
          end
        end
      end

      block = create_listing_block(
        parent,
        source.gsub("\t", '    '),
        attrs.merge({
          'style'    => 'source',
          'language' => 'go',
        })
      )
      block.title = playground_key ? "#{filename} icon:play-circle-o[title=View in Go Playground, window=_blank, link=https://play.golang.org/p/#{playground_key}]" : filename
      block.assign_caption nil
      block
    end
  end
end

# Expands to `go doc` output
#
#   godoc::go/ast.Print[]
class GoDocMacro < Asciidoctor::Extensions::BlockMacroProcessor
  use_dsl

  named :godoc

  def process(parent, target, attrs)
    m = %r<^((?:[\w.]+/)*\w+)\.([\w.]+)$>.match(target)
    opts = attrs.delete(1) || ''

    pkg, entry = m[1], m[2]
    if /^[a-z]/ === entry
      opts += ' -u'
    end
    godoc, ok = run_cached('go-doc', "go doc #{opts} #{target} 2>&1")
    unless ok
      Asciidoctor::LoggerManager.logger.warn "godoc::#{target}[] failed: #{godoc.lines.first&.chomp}"
      return create_listing_block(parent, "// go doc #{target}: 取得できませんでした", attrs.merge({ 'title' => "godoc: #{target}" }))
    end
    godoc = godoc.sub(/\Apackage .*\n\n/, '') # 新しめの go doc は先頭にパッケージ行を出す
    godoc.sub!(/\n\n\n.*$/m, '')
    decl, *doc = godoc.split(/^ {4}/)
    decl_block = create_listing_block(
      parent,
      decl.gsub("\t", '    ').lines.map(&:chomp),
      attrs.merge({
        'style'    => 'source',
        'language' => 'go',
        'title'    => "godoc: https://pkg.go.dev/#{pkg}##{entry}[#{target}]",
      })
    )
    # TODO doc
    decl_block
  end
end

class GoSourceMacro < Asciidoctor::Extensions::InlineMacroProcessor
  use_dsl

  named :gosource

  def process(parent, target, attrs)
    ref = attrs.delete(1)
    text = target.sub(%r(^.+/), '').sub('#L', ':')
    create_anchor(parent, "<code>#{text}</code>", { type: :link, target: %(https://github.com/golang/go/blob/#{ref}/#{target}) }.merge(attrs)).convert
  end
end

class TermMacro < Asciidoctor::Extensions::InlineMacroProcessor
  use_dsl

  named :term

  def process parent, target, attrs
    if parent.document.attributes['backend'] == 'pdf'
      %(#{target}（#{attrs[1]}）)
    else
      %(#{target}（<dfn>#{attrs[1]}</dfn>）)
    end
  end
end

# API が導入されたバージョン、非推奨になったバージョンを示すバッジ
#
#   since:1.23[]              => Go 1.23〜
#   since:x/tools@v0.50.0[]   => x/tools v0.50.0〜
#   deprecated:1.22[]         => Go 1.22 で非推奨
#
# Go 1.0 からあるものには書かない。ドキュメント属性 since-min（既定値 1.1）より
# 前の Go のバージョンの since はバッジにしない（:since-min: 1.18 などで調整する）。
module VersionBadge
  def self.label(target)
    if (m = /\A(x\/\w+)@(v[\d.]+)\z/.match(target))
      [m[1], m[2]]
    elsif /\A1\.\d+(\.\d+)?\z/ === target
      ['Go', target]
    else
      raise ArgumentError, "unknown version: #{target}"
    end
  end

  def self.render(parent, kind, text)
    if parent.document.basebackend?('html')
      %(<span class="version-badge version-badge-#{kind}">#{text}</span>)
    else
      "（#{text}）"
    end
  end
end

class SinceMacro < Asciidoctor::Extensions::InlineMacroProcessor
  use_dsl

  named :since

  def process(parent, target, attrs)
    mod, ver = VersionBadge.label(target)
    min = parent.document.attr('since-min', '1.1')
    return '' if mod == 'Go' && Gem::Version.new(ver) < Gem::Version.new(min)
    VersionBadge.render(parent, 'since', "#{mod} #{ver}〜")
  rescue ArgumentError => e
    Asciidoctor::LoggerManager.logger.warn "since:#{target}[]: #{e.message}"
    ''
  end
end

class DeprecatedMacro < Asciidoctor::Extensions::InlineMacroProcessor
  use_dsl

  named :deprecated

  def process(parent, target, attrs)
    mod, ver = VersionBadge.label(target)
    VersionBadge.render(parent, 'deprecated', "#{mod} #{ver} で非推奨")
  rescue ArgumentError => e
    Asciidoctor::LoggerManager.logger.warn "deprecated:#{target}[]: #{e.message}"
    ''
  end
end

Asciidoctor::Extensions.register do
  block_macro  GoExampleMacro
  block_macro  GoDocMacro
  inline_macro GoSourceMacro
  inline_macro TermMacro
  inline_macro SinceMacro
  inline_macro DeprecatedMacro
  preprocessor do
    process do |document, reader|
      document.attributes['go_version'] = GO_VERSION
      document.attributes['revnumber'] = %x(git describe --tags --always HEAD).chomp
      document.attributes['revdate'] = %x(git log -1 --pretty=%aI).chomp
      reader
    end
  end
end
