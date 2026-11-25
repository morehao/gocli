import { useCallback, useEffect, useState } from 'react'
import {
  App,
  Button,
{{- if .NeedDatePicker}}
  DatePicker,
{{- end}}
  Form,
  Input,
{{- if .NeedInputNumber}}
  InputNumber,
{{- end}}
  Modal,
{{- if .NeedSelect}}
  Select,
{{- end}}
  Space,
{{- if .NeedSwitch}}
  Switch,
{{- end}}
  Table,
  Tooltip,
  type TableProps,
} from 'antd'
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { actionColumn, tableScrollX } from '@morehao/ark-iam-react'
import { PageContainer } from '@ark-go-starter/ui'
import {
  create{{.StructName}},
  delete{{.StructName}},
  get{{.StructName}}s,
  update{{.StructName}},
} from '@ark-go-starter/api'
import type { {{.StructName}}CreateReq, {{.StructName}}PageListItem } from '@ark-go-starter/types'
{{- if .NeedFmtTime}}

/** unix 秒转本地时间字符串，空值显示 '-' */
function fmtTime(v?: number | null): string {
  if (v === undefined || v === null) return '-'
  return new Date(v * 1000).toLocaleString()
}
{{- end}}

export default function {{.StructName}}Page() {
  const { message } = App.useApp()
  const [data, setData] = useState<{{.StructName}}PageListItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<{{.StructName}}PageListItem | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm<{{.StructName}}CreateReq>()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await get{{.StructName}}s({ page, pageSize })
      setData(res.list ?? [])
      setTotal(res.total ?? 0)
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载{{.Comment}}失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, message])

  useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (record: {{.StructName}}PageListItem) => {
    setEditing(record)
    form.setFieldsValue(record as unknown as {{.StructName}}CreateReq)
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      if (editing) {
        await update{{.StructName}}(editing.{{.PKJsonName}}, values)
        message.success('更新成功')
      } else {
        await create{{.StructName}}(values)
        message.success('创建成功')
      }
      setModalOpen(false)
      load()
    } catch (e) {
      if (e instanceof Error) message.error(e.message)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (id: {{.PKTSType}}) => {
    try {
      await delete{{.StructName}}(id)
      message.success('删除成功')
      load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  const columns: TableProps<{{.StructName}}PageListItem>['columns'] = [
{{- range .Fields}}
{{- if eq .Widget "date"}}
    { title: '{{.Label}}', dataIndex: '{{.JsonName}}', key: '{{.JsonName}}', width: 200, render: (v: number | null | undefined) => fmtTime(v) },
{{- else if eq .Widget "switch"}}
    { title: '{{.Label}}', dataIndex: '{{.JsonName}}', key: '{{.JsonName}}', width: 100, render: (v?: boolean) => (v ? '是' : '否') },
{{- else if eq .Widget "number"}}
    { title: '{{.Label}}', dataIndex: '{{.JsonName}}', key: '{{.JsonName}}', width: 140 },
{{- else if .IsPK}}
    { title: '{{.Label}}', dataIndex: '{{.JsonName}}', key: '{{.JsonName}}', width: 100 },
{{- else}}
    { title: '{{.Label}}', dataIndex: '{{.JsonName}}', key: '{{.JsonName}}', width: 200, ellipsis: true },
{{- end}}
{{- end}}
    actionColumn<{{.StructName}}PageListItem>({
      variant: 'icon',
      actions: (record) => [
        { key: 'edit', label: '编辑', icon: <EditOutlined />, onClick: () => openEdit(record) },
        {
          key: 'delete',
          label: '删除',
          danger: true,
          icon: <DeleteOutlined />,
          confirm: '确认删除该{{.Comment}}？',
          onClick: () => void handleDelete(record.{{.PKJsonName}}),
        },
      ],
    }),
  ]

  return (
    <PageContainer
      title="{{.Comment}}"
      description="由 gocli 生成，数据来自 GET /v1/{{.PackageName}}/{{.ResourceKebab}}"
      extra={
        <Space>
          <Tooltip title="刷新">
            <Button icon={<ReloadOutlined />} aria-label="刷新" onClick={load} />
          </Tooltip>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
            新建{{.Comment}}
          </Button>
        </Space>
      }
    >
      <Table<{{.StructName}}PageListItem>
        rowKey="{{.PKJsonName}}"
        tableLayout="fixed"
        scroll={tableScrollX(columns)}
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={ {
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        } }
      />

      <Modal
        title={editing ? '编辑{{.Comment}}' : '新建{{.Comment}}'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={submitting}
        forceRender
      >
        <Form form={form} layout="vertical" preserve={false}>
{{- range .Fields}}
{{- if .IsPK}}{{else}}
          <Form.Item
            name="{{.JsonName}}"
            label="{{.Label}}"
{{- if .IsRequired}}
            rules={[{ required: true, message: '请输入{{.Label}}' }]}
{{- end}}
          >
{{- if eq .Widget "textarea"}}
            <Input.TextArea rows={3} placeholder="请输入{{.Label}}" />
{{- else if eq .Widget "number"}}
            <InputNumber style={ { width: '100%' } } placeholder="请输入{{.Label}}" />
{{- else if eq .Widget "switch"}}
            <Switch />
{{- else if eq .Widget "date"}}
            <DatePicker style={ { width: '100%' } } showTime />
{{- else if eq .Widget "select"}}
            <Select placeholder="请选择{{.Label}}" allowClear options={[]} />
{{- else}}
            <Input placeholder="请输入{{.Label}}" />
{{- end}}
          </Form.Item>
{{- end}}
{{- end}}
        </Form>
      </Modal>
    </PageContainer>
  )
}
